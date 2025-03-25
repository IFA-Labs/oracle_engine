package aggregator

import (
	"context"
	"oracle_engine/internal/config"
	"oracle_engine/internal/models"
	"oracle_engine/internal/utils"
)

/*
Aggegator:
Aggr unit based on each asset on the system.
Each unit can be changed dynamically to spin up multiple threads.
Each unit should handle whatever is thrown at it
*/

type AggrUnitCh chan models.UnifiedPrice

type Aggregator struct {
	InitialAggregatorUnitCount uint8
	AggegatorChannelsMap       map[string]AggrUnitCh
}

func New(cfg *config.Config) *Aggregator {
	return &Aggregator{
		// Use env vars later
		InitialAggregatorUnitCount: 1,
	}
}

func (ag *Aggregator) Start(ctx context.Context, cfg *config.Config) {
	// spin up units based on the assets available
	for _, asset := range cfg.Assets {
		// TODO: calculate asset ID using identity string to hash
		assetID := utils.GenerateIDForAsset(asset.InternalAssetIdentity)

		// create channel for the unit
		aggrUnitCh := make(chan models.UnifiedPrice, 10)
		ag.AggegatorChannelsMap[assetID] = aggrUnitCh

		assetAggregatorUnit := NewAggregatorUnit(
			aggrUnitCh,
			ag.InitialAggregatorUnitCount,
			assetID,
		)
		go assetAggregatorUnit.RunAggregatorThreadUnit(ctx)
	}
	// run the main aggregator itself
}

func (ag *Aggregator) Run(ctx context.Context, priceChan AggrUnitCh) {
	for {
		select {
		case <-ctx.Done():
			return
		case price := <-priceChan:
			// retrieve the id of the price
			assetID := price.AssetID
			// throw to the channel for that id
			ag.AggegatorChannelsMap[assetID] <- price
		}
	}
}
