package main

import (
	"context"
	"oracle_engine/internal/aggregator"
	"oracle_engine/internal/config"
	"oracle_engine/internal/consensus"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/datastream"
	"oracle_engine/internal/datastream/binance"
	"oracle_engine/internal/datastream/monierate"
	"oracle_engine/internal/datastream/pyth"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/pricepool"
	"oracle_engine/internal/relayer"
	"oracle_engine/internal/server"
	"os"
	"os/signal"
	"syscall"

	_ "oracle_engine/docs"
)

func main() {
	// Initialize Zap logger
	logging.Init()
	defer logging.Sync()

	cfg := config.Load()
	logging.Logger.Info("Starting oracle")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DB
	db, _ := timescale.NewTimescaleDB(cfg.DB_URL)

	// Initialize Data Stream
	priceCh := make(chan models.Price, 100)
	ds := datastream.New(cfg, priceCh, db)

	// Register feeds
	ds.RegisterFeed(binance.New())
	ds.RegisterFeed(pyth.New())
	ds.RegisterFeed(monierate.New(cfg))

	// Start Data Stream
	go ds.Start(ctx, cfg)

	// Price pool
	pp := pricepool.New(cfg, priceCh)
	go pp.Start(ctx)

	// Aggr
	aggr := aggregator.New(ctx, cfg)
	go aggr.Run(ctx, pp.OutChannel())

	relayer := relayer.New(cfg, db)
	consensus := consensus.New(relayer, db)
	go consensus.Ambassador(ctx, aggr.AggrOutCh)

	srv := server.New(cfg, consensus.IssuanceChan(), db)
	go srv.StartHTTPServer(ctx)

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logging.Logger.Info("Shutting down")
}
