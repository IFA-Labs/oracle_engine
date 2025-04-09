package services

import (
	"context"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/repository"
)

type IssuanceService interface {
	SaveIssuance(ctx context.Context, issuance models.Issuance) error
	GetIssuance(ctx context.Context, id string) (*models.Issuance, error)
}

type issuanceService struct {
	issuanceRepo repository.IssuanceRepository
	priceRepo    repository.PriceRepository
}

func NewIssuanceService(issuanceRepo repository.IssuanceRepository, priceRepo repository.PriceRepository) IssuanceService {
	return &issuanceService{
		issuanceRepo: issuanceRepo,
		priceRepo:    priceRepo,
	}
}

func (s *issuanceService) SaveIssuance(ctx context.Context, issuance models.Issuance) error {
	return s.issuanceRepo.SaveIssuance(ctx, issuance)
}

func (s *issuanceService) GetIssuance(ctx context.Context, id string) (*models.Issuance, error) {
	return s.issuanceRepo.GetIssuance(ctx, id)
}
