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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	cfg     *config.Config
	priceCh chan models.Issuance
	db      *timescale.TimescaleDB
	gormDB  *timescale.TimescaleGORM
	api     *api.API
}

func New(cfg *config.Config, priceCh chan models.Issuance, db *timescale.TimescaleDB) *Server {
	// Initialize GORM DB for dashboard operations
	gormDB, err := timescale.NewTimescaleGORM(cfg.DB_URL)
	if err != nil {
		logging.Logger.Fatal("Failed to initialize GORM DB", zap.Error(err))
	}

	// Initialize repositories
	priceRepo := repository.NewPriceRepository(db)
	issuanceRepo := repository.NewIssuanceRepository(db)
	dashboardRepo := repository.NewDashboardRepository(gormDB.GetDB())

	// Initialize services
	priceService := services.NewPriceService(priceRepo)
	issuanceService := services.NewIssuanceService(issuanceRepo, priceRepo)
	dashboardService := services.NewDashboardService(dashboardRepo, cfg.JWTSecret, cfg)

	// Initialize API
	api := api.NewAPI(priceService, issuanceService, dashboardService, priceCh, cfg)

	return &Server{
		cfg:     cfg,
		priceCh: priceCh,
		db:      db,
		gormDB:  gormDB,
		api:     api,
	}
}

func (s *Server) StartHTTPServer(ctx context.Context) {
	// Create router with middleware
	router := gin.Default()

	// Register routes
	s.api.RegisterRoutes(router)

	// Apply middleware
	handler := middleware.CORS(router)
	handler = middleware.Logging(handler)
	handler = middleware.Recovery(handler)
	port := s.cfg.SERVER_PORT
	if port == "" {
		port = "5001"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		logging.Logger.Info("Starting HTTP server on :", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logging.Logger.Info("Shutting down server...")
	server.Shutdown(context.Background())
	s.Close()
}

// Close properly closes database connections
func (s *Server) Close() {
	if s.gormDB != nil {
		if err := s.gormDB.Close(); err != nil {
			logging.Logger.Error("Failed to close GORM DB", zap.Error(err))
		}
	}
	// Note: The original TimescaleDB doesn't have a Close method in the provided code
	// You might want to add one if needed
}
