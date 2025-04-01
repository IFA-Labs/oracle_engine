package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"oracle_engine/internal/config"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/utils"

	"go.uber.org/zap"
)

type IssuanceChan chan models.Issuance

type Server struct {
	cfg     *config.Config
	priceCh IssuanceChan // From Consensus to SSE
	db      *timescale.TimescaleDB
}

func New(cfg *config.Config, priceCh IssuanceChan, db *timescale.TimescaleDB) *Server {
	return &Server{
		cfg:     cfg,
		priceCh: priceCh,
		db:      db,
	}
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (adjust for production)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight (OPTIONS) requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to next handler
		next.ServeHTTP(w, r)
	})
}

func (s *Server) StartHTTPServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/assets", s.handleAssets)
	mux.HandleFunc("/last-price", s.handleLastPrice)
	mux.HandleFunc("/recent-prices", s.handleRecentPrices)
	mux.Handle("/", http.FileServer(http.Dir("./web"))) // Serve static files

	handler := corsMiddleware(mux)

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

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.cfg.Assets)
}

func (s *Server) handleLastPrice(w http.ResponseWriter, r *http.Request) {
	asset := r.URL.Query().Get("asset")
	if asset == "" {
		// Return all assets' last prices if no specific asset is provided
		prices := make(map[string]float64)
		for _, a := range s.cfg.Assets {
			price, err := s.db.GetLastPrice(r.Context(), utils.GenerateIDForAsset(a.InternalAssetIdentity))
			if err != nil {
				logging.Logger.Error("Failed to fetch last price", zap.String("asset", a.Name), zap.Error(err))
				continue
			}
			prices[utils.GenerateIDForAsset(a.InternalAssetIdentity)] = price.Number() // Convert to token units
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prices)
		return
	}

	// Single asset case
	price, err := s.db.GetLastPrice(r.Context(), asset)
	if err != nil {
		http.Error(w, "Failed to fetch last price", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{asset: price.Number()})
}

func (s *Server) handleRecentPrices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case price := <-s.priceCh:
			data, err := json.Marshal(price)
			if err != nil {
				logging.Logger.Error("Failed to marshal price", zap.Error(err))
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
