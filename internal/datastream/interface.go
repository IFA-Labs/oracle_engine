package datastream

import (
	"context"
	"oracle_engine/internal/models"
	"time"
)

type PriceFeed interface {
	FetchPrice(ctx context.Context, baseAssetID string, quoteAssetID string, internalAssetId string) (*models.Price, error)
	Name() string
	Interval() time.Duration
	AssetID() string
}
