package datastream

import (
	"context"
	"oracle_engine/internal/models"
	"time"
)

type PriceFeed interface {
	FetchPrice(ctx context.Context, assetID string) (*models.Price, error)
	Name() string
	Interval() time.Duration
	AssetID() string
}
