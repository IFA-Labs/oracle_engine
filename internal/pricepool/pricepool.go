package pricepool

import (
	"context"
	"encoding/json"
	"errors"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/pricepool/dlq"
	"oracle_engine/internal/pricepool/outlier"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type PricePool struct {
	client   *redis.Client
	cfg      *config.Config
	out      chan models.UnifiedPrice // To Aggregators (future)
	dlq      *dlq.DLQ                 // Dead-letter queue
	incoming chan models.Price        // From Data Stream
}

func New(cfg *config.Config, incoming chan models.Price) *PricePool {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost" // Default for non-Docker
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "", // Add via env if needed
		DB:       0,
	})

	return &PricePool{
		client:   client,
		cfg:      cfg,
		incoming: incoming,
		out:      make(chan models.UnifiedPrice, 100),
		dlq:      dlq.NewDLQ(),
	}
}

func (p *PricePool) Start(ctx context.Context) {
	// Handle incoming prices
	go p.processIncoming(ctx)

	// Periodic cleanup
	go p.cleanup(ctx)
}

func (p *PricePool) processIncoming(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case price := <-p.incoming:
			if err := p.validateAndStore(ctx, price); err != nil {
				p.dlq.Enqueue(price, err)
				logging.Logger.Warn("Invalid price sent to DLQ",
					zap.Any("price", price),
					zap.Error(err))
				continue
			}
			p.out <- price.ToUnified() // Pass to Aggregators
			logging.Logger.Info("Price stored",
				zap.String("asset", price.Asset),
				zap.Float64("value", price.ToUnified().Number()))
		}
	}
}

func (p *PricePool) validateAndStore(ctx context.Context, price models.Price) error {
	// Basic validation
	if price.Value <= 0 || price.Asset == "" {
		return errors.New("invalid price: negative value or missing asset")
	}

	// Serialize price
	data, err := json.Marshal(price)
	if err != nil {
		return err
	}

	// Store in Redis with TTL
	key := "pricepool:" + price.Asset
	if err := p.client.RPush(ctx, key, data).Err(); err != nil {
		return err
	}
	p.client.Expire(ctx, key, time.Duration(p.cfg.PricePoolTTL)*time.Minute)

	return nil
}

func (p *PricePool) cleanup(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Configurable
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.filterOutliers(ctx)
		}
	}
}

func (p *PricePool) filterOutliers(ctx context.Context) {
	for _, asset := range p.cfg.Assets {
		key := "pricepool:" + asset.Name
		prices, err := p.getPrices(ctx, key)
		if err != nil {
			logging.Logger.Error("Failed to fetch prices for cleanup", zap.Error(err))
			continue
		}

		filtered := outlier.FilterOutliers(prices)
		if len(filtered) < len(prices) {
			logging.Logger.Info("Outliers removed",
				zap.String("asset", asset.Name),
				zap.Int("removed", len(prices)-len(filtered)))
			p.updatePrices(ctx, key, filtered)
		}
	}
}

func (p *PricePool) getPrices(ctx context.Context, key string) ([]models.Price, error) {
	vals, err := p.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var prices []models.Price
	for _, val := range vals {
		var price models.Price
		if err := json.Unmarshal([]byte(val), &price); err != nil {
			continue // Skip malformed entries
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (p *PricePool) updatePrices(ctx context.Context, key string, prices []models.Price) {
	p.client.Del(ctx, key) // Clear old list
	for _, price := range prices {
		data, _ := json.Marshal(price)
		p.client.RPush(ctx, key, data)
	}
	p.client.Expire(ctx, key, time.Duration(p.cfg.PricePoolTTL)*time.Minute)
}

func (p *PricePool) OutChannel() chan models.UnifiedPrice {
	return p.out
}
