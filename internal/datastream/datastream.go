package datastream

import (
	"context"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"

	"go.uber.org/zap"
)

type DataStream struct {
	feeds map[string]PriceFeed
	out   chan models.Price
}

func New(cfg *config.Config, out chan models.Price) *DataStream {
	feeds := make(map[string]PriceFeed)
	return &DataStream{feeds: feeds, out: out}
}

func (ds *DataStream) RegisterFeed(feed PriceFeed) {
	// will include a string generator
	ds.feeds[feed.Name()] = feed
}

// A job scheduler here that runs at intervals based on the feed
// TODO: restructure so asset is based on feed instead
func (ds *DataStream) Start(ctx context.Context, cfg *config.Config) {
	for _, asset := range cfg.Assets {
		for _, feedCfg := range asset.Feeds {
			feed := ds.feeds[feedCfg.Name]
			feedAssetID := feedCfg.AssetID
			feedInternalAssetID := asset.InternalAssetIdentity
			// if feed doesn't exist, just move on meaning the asset doesn't support feed
			if feed == nil {
				// logging.Logger.Warn("Unknown feed", zap.String("name", feedCfg.Name))
				continue
			}
			go ds.runFeed(
				ctx, asset.Name, feedAssetID, feedInternalAssetID,
				feed, time.Duration(feedCfg.Interval)*time.Second)
		}
	}
}

func (ds *DataStream) runFeed(ctx context.Context, asset string, assetId string,
	internalAssetIdentity string, feed PriceFeed, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			price, err := feed.FetchPrice(ctx, assetId, internalAssetIdentity)
			if err != nil {
				logging.Logger.Error("Fetch failed",
					zap.String("feed", feed.Name()),
					zap.String("asset", asset),
					zap.Error(err))
				continue
			}
			price.Asset = asset
			if feed.Name() != "pyth" {
				continue
			}
			ds.out <- *price
			logging.Logger.Info("Price fetched",
				zap.String("asset", price.Asset),
				zap.Float64("value", price.Value),
				zap.String("source", price.Source))
		}
	}
}
