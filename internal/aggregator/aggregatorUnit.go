package aggregator

import (
	"context"
	"math"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"sync"
	"time"

	"go.uber.org/zap"
)

const BUFFER_MAX_SIZE = 10

type AggregatorUnit struct {
	ActiveThreads uint8
	AggrDevPerc   float32
	AssetID       string
	ch            AggrUnitCh
	outCh         *AggrUnitCh
	wg            sync.WaitGroup
}

func NewAggregatorUnit(
	ch AggrUnitCh,
	outCh *AggrUnitCh,
	aggrDevPerc float32,
	initialThreadCount uint8,
	assetID string,
) *AggregatorUnit {
	return &AggregatorUnit{
		ch:            ch,
		ActiveThreads: initialThreadCount,
		AssetID:       assetID,
		outCh:         outCh,
		AggrDevPerc:   aggrDevPerc,
	}
}

func (au *AggregatorUnit) RunAggregatorThreadUnit(ctx context.Context) {
	priceBuf := make([]models.UnifiedPrice, 10)
	// lock threads
	for {
		select {
		case <-ctx.Done():
			au.wg.Wait() // case killed, at least throw
			return
		case price := <-au.ch:
			logging.Logger.Warn("---yy-- ret", zap.Any("k", au.AssetID))
			priceBuf = append(priceBuf, price)
			if len(priceBuf) >= BUFFER_MAX_SIZE {
				// Batch up and price out
				au.wg.Add(1)

				copiedPrices := make([]models.UnifiedPrice, 10)
				copy(copiedPrices, priceBuf)
				go func(cp []models.UnifiedPrice) {
					defer au.wg.Done()

					threadUnitCalculateBatchAverage(
						cp, au.outCh, au.AggrDevPerc,
					)
				}(copiedPrices)
				// reset price buf
				priceBuf = []models.UnifiedPrice{}
			}
		}
	}
}

func threadUnitCalculateBatchAverage(
	batch []models.UnifiedPrice,
	outgoingCh *AggrUnitCh,
	aggr_dev_perc float32,
) {
	firstPrice := batch[0]
	avg := (firstPrice.Value + batch[len(batch)-1].Value) / 2

	sum := 0.0
	for _, p := range batch {
		pn := p.Value
		devPerc := math.Abs(pn-avg) / avg
		if devPerc > float64(aggr_dev_perc) {
			continue
		}
		sum += pn
	}
	avg = sum / float64(len(batch))

	logging.Logger.Warn("---compute babe-- ret", zap.Any("k", avg))
	// some other calc
	// source aggr and hash gen
	avgPrice := models.UnifiedPrice{
		Value:     avg,
		AssetID:   firstPrice.AssetID,
		Expo:      firstPrice.Expo, // still -18
		Timestamp: time.Now(),
		Source:    "ifa_labs",
		ReqHash:   "todo",
	}
	*outgoingCh <- avgPrice
}
