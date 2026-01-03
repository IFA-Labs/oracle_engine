package services

import (
	"context"
	"time"

	"oracle_engine/internal/models"
	"oracle_engine/internal/server/repository"
)

type PriceService interface {
	GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error)
	GetHistoricalPrice(ctx context.Context, assetID string, lookback time.Duration) (*models.UnifiedPrice, error)
	SavePrice(ctx context.Context, price models.UnifiedPrice) error
	AuditPrice(ctx context.Context, assetID string) (*models.PriceAudit, error)
	AuditPriceRange(ctx context.Context, fromTime, toTime time.Time, assetID string, limit, offset int) ([]*models.PriceAudit, error)
}

type priceService struct {
	priceRepo repository.PriceRepository
}

func NewPriceService(priceRepo repository.PriceRepository) PriceService {
	return &priceService{priceRepo: priceRepo}
}

func (s *priceService) GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error) {
	return s.priceRepo.GetLastPrice(ctx, assetID)
}

func (s *priceService) GetHistoricalPrice(ctx context.Context, assetID string, lookback time.Duration) (*models.UnifiedPrice, error) {
	return s.priceRepo.GetHistoricalPrice(ctx, assetID, lookback)
}

func (s *priceService) SavePrice(ctx context.Context, price models.UnifiedPrice) error {
	return s.priceRepo.SavePrice(ctx, price)
}

func (s *priceService) AuditPrice(ctx context.Context, assetID string) (*models.PriceAudit, error) {
	return s.priceRepo.AuditPrice(ctx, assetID)
}

func (s *priceService) AuditPriceRange(ctx context.Context, fromTime, toTime time.Time, assetID string, limit, offset int) ([]*models.PriceAudit, error) {
	return s.priceRepo.AuditPriceRange(ctx, fromTime, toTime, assetID, limit, offset)
}
