package aggregator

import (
	"context"
	"log"
	"oracle_engine/internal/models"
)

type AggregatorUnit struct {
	ActiveThreads uint8
	AssetID       string
	ch            AggrUnitCh
}

func NewAggregatorUnit(ch AggrUnitCh, initialThreadCount uint8, assetID string) *AggregatorUnit {
	return &AggregatorUnit{
		ch:            ch,
		ActiveThreads: initialThreadCount,
		AssetID:       assetID,
	}
}

func (au *AggregatorUnit) RunAggregatorThreadUnit(ctx context.Context) {
	priceBuf := make([]models.UnifiedPrice, 10)
	for {
		select {
		case <-ctx.Done():
			return
		case price := <- au.ch:
			log.Printf("%v", price.Number())
		}
	}
}
