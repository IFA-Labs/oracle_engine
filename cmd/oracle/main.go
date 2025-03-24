package main

import (
	"context"
	"oracle_engine/internal/config"
	"oracle_engine/internal/datastream"
	"oracle_engine/internal/datastream/binance"
	"oracle_engine/internal/datastream/pyth"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
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

	// Register feeds
	ds.RegisterFeed(binance.New())
	ds.RegisterFeed(pyth.New())
	// ds.RegisterFeed(coinbase.New())
	// ds.RegisterFeed(kraken.New())

	// Start Data Stream
	go ds.Start(ctx, cfg)

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logging.Logger.Info("Shutting down")
}
