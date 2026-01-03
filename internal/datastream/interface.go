package datastream

import (
	"context"
	"time"

	"oracle_engine/internal/models"
)

type PriceFeed interface {
	FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error)
	Name() string
	Interval() time.Duration
	AssetID() string
}
