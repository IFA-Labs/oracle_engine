package repository

import (
	"context"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/models"
)

type IssuanceRepository interface {
	SaveIssuance(ctx context.Context, issuance models.Issuance) error
	GetIssuance(ctx context.Context, id string) (*models.Issuance, error)
}

type issuanceRepository struct {
	db *timescale.TimescaleDB
}

func NewIssuanceRepository(db *timescale.TimescaleDB) IssuanceRepository {
	return &issuanceRepository{db: db}
}

func (r *issuanceRepository) SaveIssuance(ctx context.Context, issuance models.Issuance) error {
	return r.db.SaveIssuance(ctx, issuance)
}

func (r *issuanceRepository) GetIssuance(ctx context.Context, id string) (*models.Issuance, error) {
	return r.db.GetIssuance(ctx, id)
}
