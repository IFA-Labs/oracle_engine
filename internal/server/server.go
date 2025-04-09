package server

import (
	"context"
	"net/http"
	"oracle_engine/internal/config"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/api"
	"oracle_engine/internal/server/middleware"
	"oracle_engine/internal/server/repository"
	"oracle_engine/internal/server/services"

	"go.uber.org/zap"
)

type Server struct {
	cfg     *config.Config
	priceCh chan models.Issuance
	db      *timescale.TimescaleDB
	api     *api.API
}

func New(cfg *config.Config, priceCh chan models.Issuance, db *timescale.TimescaleDB) *Server {
	// Initialize repositories
	priceRepo := repository.NewPriceRepository(db)
	issuanceRepo := repository.NewIssuanceRepository(db)

	// Initialize services
	priceService := services.NewPriceService(priceRepo)
	issuanceService := services.NewIssuanceService(issuanceRepo, priceRepo)

	// Initialize API
	api := api.NewAPI(priceService, issuanceService, priceCh, cfg)

	return &Server{
		cfg:     cfg,
		priceCh: priceCh,
		db:      db,
		api:     api,
	}
}

func (s *Server) StartHTTPServer(ctx context.Context) {
	// Create router with middleware
	router := http.NewServeMux()

	// Register routes
	s.api.RegisterRoutes(router)

	// Apply middleware
	handler := middleware.CORS(router)
	handler = middleware.Logging(handler)
	handler = middleware.Recovery(handler)

	server := &http.Server{
		Addr:    ":5001",
		Handler: handler,
	}

	go func() {
		logging.Logger.Info("Starting HTTP server on :5001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	server.Shutdown(context.Background())
}
