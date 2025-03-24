package binance

import (
	"context"
	"oracle_engine/internal/models"
	"time"
)

type BinanceFeed struct {
	interval time.Duration // Configured via YAML
}

func New() *BinanceFeed {
	return &BinanceFeed{}
}

func (b *BinanceFeed) FetchPrice(ctx context.Context) (*models.Price, error) {
	// TODO: Replace with real Binance API call
	return &models.Price{
		Value:     50000.0, // Dummy
		Timestamp: time.Now(),
		Source:    b.Name(),
	}, nil
}

func (b *BinanceFeed) Name() string {
	return "binance"
}

func (b *BinanceFeed) AssetID() string {
	return ""
}

func (b *BinanceFeed) Interval() time.Duration {
	return b.interval // Default, overridden by config.yaml
}
