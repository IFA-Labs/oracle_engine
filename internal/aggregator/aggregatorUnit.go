package aggregator

import "context"

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
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
