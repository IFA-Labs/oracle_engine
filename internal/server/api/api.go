package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oracle_engine/internal/config"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/services"
	"oracle_engine/internal/utils"
	"strings"

	"go.uber.org/zap"
)

// @title Oracle Engine API
// @version 1.0
// @description API for accessing oracle price data and issuances
// @host localhost:5001
// @BasePath /api

type API struct {
	priceService    services.PriceService
	issuanceService services.IssuanceService
	priceCh         chan models.Issuance
	cfg             *config.Config
}

func NewAPI(priceService services.PriceService, issuanceService services.IssuanceService, priceCh chan models.Issuance, cfg *config.Config) *API {
	return &API{
		priceService:    priceService,
		issuanceService: issuanceService,
		priceCh:         priceCh,
		cfg:             cfg,
	}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	// Price endpoints
	mux.HandleFunc("/api/prices/last", a.handleLastPrice)
	mux.HandleFunc("/api/prices/stream", a.handlePriceStream)

	// Issuance endpoints
	mux.HandleFunc("/api/issuances", a.handleIssuances)
	mux.HandleFunc("/api/issuances/", a.handleIssuance)

	// Asset endpoints
	mux.HandleFunc("/api/assets", a.handleAssets)

	// Price audit endpoints
	mux.HandleFunc("/api/prices/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/audit") {
			a.handleAuditPrice(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Swagger docs
	mux.HandleFunc("/api/swagger/*", a.handleSwagger)
	mux.HandleFunc("/api/swagger.json", a.handleSwagger)
}

// @Summary Get last price for an asset
// @Description Returns the last known price for a specific asset or all assets
// @Tags prices
// @Accept json
// @Produce json
// @Param asset query string false "Asset ID to get price for"
// @Success 200 {object} map[string]float64
// @Router /prices/last [get]
func (a *API) handleLastPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := r.URL.Query().Get("asset")
	if asset == "" {
		// Return all assets' last prices
		prices := make(map[string]float64)
		for _, assetConfig := range a.cfg.Assets {
			price, err := a.priceService.GetLastPrice(r.Context(), utils.GenerateIDForAsset(assetConfig.InternalAssetIdentity))
			if err != nil {
				zap.L().Error("Failed to fetch last price", zap.String("asset", assetConfig.Name), zap.Error(err))
				continue
			}
			prices[utils.GenerateIDForAsset(assetConfig.InternalAssetIdentity)] = price.Number()
		}
		json.NewEncoder(w).Encode(prices)
		return
	}

	// Single asset case
	price, err := a.priceService.GetLastPrice(r.Context(), asset)
	if err != nil {
		http.Error(w, "Failed to fetch last price", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]float64{asset: price.Number()})
}

// @Summary Stream price updates
// @Description Server-Sent Events stream of price updates
// @Tags prices
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Router /prices/stream [get]
func (a *API) handlePriceStream(w http.ResponseWriter, r *http.Request) {
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
		case price := <-a.priceCh:
			data, err := json.Marshal(price)
			if err != nil {
				zap.L().Error("Failed to marshal price", zap.Error(err))
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (a *API) handleIssuances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var issuance models.Issuance
		if err := json.NewDecoder(r.Body).Decode(&issuance); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := a.issuanceService.SaveIssuance(r.Context(), issuance); err != nil {
			http.Error(w, "Failed to save issuance", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(issuance)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary Get issuance details
// @Description Returns details of a specific issuance
// @Tags issuances
// @Accept json
// @Produce json
// @Param id path string true "Issuance ID"
// @Success 200 {object} models.Issuance
// @Router /issuances/{id} [get]
func (a *API) handleIssuance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/issuances/"):]
	if id == "" {
		http.Error(w, "Issuance ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		issuance, err := a.issuanceService.GetIssuance(r.Context(), id)
		if err != nil {
			http.Error(w, "Failed to get issuance", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(issuance)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary Get available assets
// @Description Returns list of all available assets
// @Tags assets
// @Produce json
// @Success 200 {array} config.AssetConfig
// @Router /assets [get]
func (a *API) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(a.cfg.Assets)
}

// @Summary Get price audit
// @Description Returns audit information for a specific price
// @Tags prices
// @Accept json
// @Produce json
// @Param id path string true "Price ID"
// @Success 200 {object} map[string]interface{}
// @Router /prices/{id}/audit [get]
func (a *API) handleAuditPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/prices/") || !strings.HasSuffix(path, "/audit") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/prices/"), "/audit")
	if id == "" {
		http.Error(w, "Price ID required", http.StatusBadRequest)
		return
	}

	priceAudit, err := a.priceService.AuditPrice(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to audit price, %v", err), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"id":            priceAudit.AggregatedPrice.ID,
		"asset_id":      priceAudit.AssetID,
		"price":         priceAudit.AggregatedPrice,
		"raw_prices":    priceAudit.RawPrices,
		"raw_count":     len(priceAudit.RawPrices),
		"internalAsset": priceAudit.AggregatedPrice.Number(),
	}
	json.NewEncoder(w).Encode(resp)
}
