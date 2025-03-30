package main

import (
	"context"
	"oracle_engine/internal/aggregator"
	"oracle_engine/internal/config"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/datastream"
	"oracle_engine/internal/datastream/binance"
	"oracle_engine/internal/datastream/pyth"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/pricepool"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Initialize Zap logger
	logging.Init()
	defer logging.Sync()

	cfg := config.Load()
	logging.Logger.Info("Starting oracle", zap.Any("config", cfg))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Data Stream
	priceCh := make(chan models.Price, 100)
	ds := datastream.New(cfg, priceCh)
	db, _ := timescale.NewTimescaleDB("postgres://user:pass@timescale:5432/oracle")

	l, err := db.GetLastPrice(ctx, "0x1234")
	if err != nil {
		logging.Logger.Warn("----- Wrong", zap.Any("rec", err))
	} else {

		logging.Logger.Warn("----- Definitely", zap.Any("rec", l))
	}

	// Register feeds
	ds.RegisterFeed(binance.New())
	ds.RegisterFeed(pyth.New())
	// ds.RegisterFeed(coinbase.New())
	// ds.RegisterFeed(kraken.New())

	// Start Data Stream
	go ds.Start(ctx, cfg)

	// Price pool
	pp := pricepool.New(cfg, priceCh)
	go pp.Start(ctx)

	// Aggr
	aggr := aggregator.New(ctx, cfg)
	aggr.Run(ctx, pp.OutChannel())

	go ll(ctx, aggr)

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logging.Logger.Info("Shutting down")
}

func ll(ctx context.Context, aggr *aggregator.Aggregator) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-aggr.OutCh():
			logging.Logger.Sugar().Info(
				" ---------- For Consensus",
				zap.Any("name", p.AssetID),
				zap.Any("avg", p.Normalize()),
			)
		}
	}
}
