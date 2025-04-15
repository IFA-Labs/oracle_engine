package relayer

import (
	"context"
	"oracle_engine/internal/config"
	"oracle_engine/internal/models"
)

// / Relayer is a service that takes issuances requests
// / and sends to the contract
// / It also updates the status of the issuance request
// / in local db (also the asset instance)
type Relayer struct {
	cfg                 *config.Config
	assetToRoutineChMap map[string]chan *models.Issuance
}

func New(config *config.Config) *Relayer {
	return &Relayer{
		cfg:                 config,
		assetToRoutineChMap: make(map[string]chan *models.Issuance),
	}
}

// / Start treat latest issuance with utmost priority
// / Start a go routine for each issuance
// / Each contract has its own go routine
func (r *Relayer) Start(ctx context.Context, issuanceCh chan *models.Issuance) error {
	for _, asset := range r.cfg.Assets {
		r.assetToRoutineChMap[asset.InternalAssetIdentity] = make(chan *models.Issuance)
		go r.startRoutine(ctx, asset.InternalAssetIdentity)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case issuance := <-issuanceCh:
			r.assetToRoutineChMap[issuance.Price.AssetID] <- issuance
		}
	}
}

func (r *Relayer) startRoutine(ctx context.Context, assetID string) {
	ch := r.assetToRoutineChMap[assetID]
	for issuance := range ch {
		for _, ctrct := range r.cfg.Contracts {
			go r.ConveyIssuanceToContract(ctx, issuance, ctrct)
		}
	}
}

func (r *Relayer) ConveyIssuanceToContract(ctx context.Context, issuance *models.Issuance, ctrct config.ContractConfig) error {
	return nil
}
