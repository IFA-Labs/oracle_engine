package consensus

import (
	"context"
	"oracle_engine/internal/aggregator"
	"oracle_engine/internal/consensus/voting/weighted"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/relayer"

	"github.com/google/uuid"

	"go.uber.org/zap"
)

// This is an internal consensus layer
// Fetches aggregated prices and validates
// Based on last stored price, for now,
// Just on a single price change, determine
// If outrageous, throw to store as invalid
// Else if too close, discard
// Else issue relay request to relayer
type Consensus struct {
	// out channel
	relayer    relayer.Relayer
	db         timescale.TimescaleDB
	issuanceCh chan models.Issuance
}

func New(relayer *relayer.Relayer, db *timescale.TimescaleDB) *Consensus {

	logging.Logger.Warn("---yy-- home run")
	return &Consensus{
		relayer:    *relayer,
		db:         *db,
		issuanceCh: make(chan models.Issuance, 10),
	}
}

func (c *Consensus) IssuanceChan() chan models.Issuance {
	return c.issuanceCh
}

func (c *Consensus) Ambassador(ctx context.Context, incomingCh aggregator.AggrUnitCh) {
	tmpIssuanceCh := make(chan models.Issuance, 10)

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-incomingCh:
			logging.Logger.Warn(
				" ---------- For Consensus",
				zap.Any("name", p.AssetID),
				zap.Any("avg", p.Number()),
			)
			tmpIssuanceCh <- c.processAggrPrice(ctx, p)
		case issuance := <-tmpIssuanceCh:
			c.handleIssuance(ctx, issuance)
		}
	}
}

func (c *Consensus) handleIssuance(ctx context.Context, issuance models.Issuance) {
	logging.Logger.Info("Issuance", zap.Int("num", int(issuance.Price.Number())))
	c.db.SaveIssuance(ctx, issuance)
	c.issuanceCh <- issuance
}

func (c *Consensus) processAggrPrice(
	ctx context.Context,
	price models.UnifiedPrice,
) models.Issuance {
	id := uuid.NewString()
	// TODO: fetch more prices from db (last n prices)
	lastPrice, err := c.db.GetLastPrice(ctx, price.AssetID)
	if err != nil {
		// handle error
		logging.Logger.Error("Couldn't fetch last price for consensus", zap.Any("id", price.AssetID))
		lastPrice = &price
	}
	lastPrices := []models.UnifiedPrice{*lastPrice}
	issuance := weighted.CalculateWeightedAveragePrice(id, price, lastPrices)

	// Save the aggregated price in price and link
	// the batch through ids
	c.db.LinkRawPricesToAggregatedPrice(
		ctx,
		issuance.Price.ID,
		issuance.PriceTimestamp,
		price.ConnectedPriceIDs,
	)

	return issuance
}
