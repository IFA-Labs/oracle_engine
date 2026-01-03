package repository

import (
	"context"
	"time"

	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/models"
)

type PriceRepository interface {
	SavePrice(ctx context.Context, price models.UnifiedPrice) error
	GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error)
	GetHistoricalPrice(ctx context.Context, assetID string, lookback time.Duration) (*models.UnifiedPrice, error)
	AuditPrice(ctx context.Context, assetID string) (*models.PriceAudit, error)
	AuditPriceRange(ctx context.Context, fromTime, toTime time.Time, assetID string, limit, offset int) ([]*models.PriceAudit, error)
}

type priceRepository struct {
	db *timescale.TimescaleDB
}

func NewPriceRepository(db *timescale.TimescaleDB) PriceRepository {
	return &priceRepository{db: db}
}

func (r *priceRepository) SavePrice(ctx context.Context, price models.UnifiedPrice) error {
	return r.db.SavePrice(ctx, price)
}

func (r *priceRepository) GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error) {
	return r.db.GetLastPrice(ctx, assetID)
}

func (r *priceRepository) GetHistoricalPrice(ctx context.Context, assetID string, lookback time.Duration) (*models.UnifiedPrice, error) {
	return r.db.GetHistoricalPrice(ctx, assetID, lookback)
}

func (r *priceRepository) AuditPrice(ctx context.Context, assetID string) (*models.PriceAudit, error) {
	return r.db.AuditPrice(ctx, assetID)
}

func (r *priceRepository) AuditPriceRange(ctx context.Context, fromTime, toTime time.Time, assetID string, limit, offset int) ([]*models.PriceAudit, error) {
	return r.db.AuditPriceRange(ctx, fromTime, toTime, assetID, limit, offset)
}
