package services

import (
	"context"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/repository"
)

type PriceService interface {
	GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error)
	SavePrice(ctx context.Context, price models.UnifiedPrice) error
	AuditPrice(ctx context.Context, assetID string) (*models.PriceAudit, error)
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

func (s *priceService) SavePrice(ctx context.Context, price models.UnifiedPrice) error {
	return s.priceRepo.SavePrice(ctx, price)
}

func (s *priceService) AuditPrice(ctx context.Context, assetID string) (*models.PriceAudit, error) {
	return s.priceRepo.AuditPrice(ctx, assetID)
}
