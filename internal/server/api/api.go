package api

import (
	"fmt"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/services"
	"oracle_engine/internal/utils"

	"go.uber.org/zap"

	_ "oracle_engine/docs"

	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Oracle Engine API
// @version 1.0
// @description IFA LABS Oracle Engine API provides real-time, reliable asset prices using an aggregated moving window algorithm to ensure stability and reduce manipulation.
// @host localhost:8000
// @host 146.190.186.116:8000
// @BasePath /api
// @contact.name   IfaLabs
// @contact.url     https://ifalabs.com
// @contact.email  ifalabstudio@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
type API struct {
	priceService    services.PriceService
	issuanceService services.IssuanceService
	priceCh         chan models.Issuance
	priceStreamer   *PriceStreamer
	cfg             *config.Config
}

func NewAPI(priceService services.PriceService, issuanceService services.IssuanceService, priceCh chan models.Issuance, cfg *config.Config) *API {

	priceStreamer := NewPriceStreamer(priceCh, logging.Logger)
	priceStreamer.Start()

	return &API{
		priceService:    priceService,
		issuanceService: issuanceService,
		priceCh:         priceCh,
		priceStreamer:   priceStreamer,
		cfg:             cfg,
	}
}

func (a *API) RegisterRoutes(router *gin.Engine) {
	// Price endpoints
	router.GET("/api/prices/last", a.handleLastPrice)
	// router.GET("/api/prices/stream", a.handlePriceStream)
	router.GET("/api/prices/stream", a.priceStreamer.HandleStream)

	// Issuance endpoints
	router.POST("/api/issuances", a.handleIssuances)
	router.GET("/api/issuances/:id", a.handleIssuance)

	// Asset endpoints
	router.GET("/api/assets", a.handleAssets)

	// Price audit endpoint
	router.GET("/api/prices/:id/audit", a.handleAuditPrice)

	url := ginSwagger.URL("/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Health check endpoint
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	},
	)
}

// @Summary Get last price for an asset
// @Description Returns the last known price for a specific asset or all assets
// @Tags prices
// @Accept json
// @Produce json
// @Param asset query string false "Asset ID to get price for"
// @Param changes query string false "Comma-separated list of price change periods (e.g. '7d,3d,24h'). Default is '7d'"
// @Success 200 {object} map[string]models.UnifiedPrice
// @Router /prices/last [get]
func (a *API) handleLastPrice(c *gin.Context) {
	asset := c.Query("asset")
	changesParam := c.DefaultQuery("changes", "7d") // Default to 7d if not specified

	// Parse change periods
	changePeriods := strings.Split(changesParam, ",")
	periodDurations := make(map[string]time.Duration)

	for _, period := range changePeriods {
		period = strings.TrimSpace(period)
		if period == "" {
			continue
		}

		// Parse period string (e.g. "7d", "24h")
		var duration time.Duration

		if strings.HasSuffix(period, "d") {
			days, err := strconv.Atoi(strings.TrimSuffix(period, "d"))
			if err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid period format: %s", period)})
				return
			}
			duration = time.Duration(days) * 24 * time.Hour
		} else if strings.HasSuffix(period, "h") {
			hours, err := strconv.Atoi(strings.TrimSuffix(period, "h"))
			if err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid period format: %s", period)})
				return
			}
			duration = time.Duration(hours) * time.Hour
		} else {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Unsupported period format: %s", period)})
			return
		}

		periodDurations[period] = duration
	}

	if asset == "" {
		// Return all assets' last prices
		prices := make(map[string]*models.UnifiedPrice)
		for _, assetConfig := range a.cfg.Assets {
			assetID := utils.GenerateIDForAsset(assetConfig.InternalAssetIdentity)
			price, err := a.priceService.GetLastPrice(c.Request.Context(), assetID)
			if err != nil {
				zap.L().Error("Failed to fetch last price", zap.String("asset", assetConfig.Name), zap.Error(err))
				continue
			}

			// Calculate price changes for each period
			price.PriceChanges = make([]models.PriceChange, 0, len(periodDurations))
			for period, duration := range periodDurations {
				historicalPrice, err := a.priceService.GetHistoricalPrice(c.Request.Context(), assetID, duration)
				if err != nil {
					zap.L().Error("Failed to fetch historical price",
						zap.String("asset", assetConfig.Name),
						zap.String("period", period),
						zap.Error(err))
					continue
				}

				if change := models.CalculatePriceChange(price, historicalPrice, period); change != nil {
					price.PriceChanges = append(price.PriceChanges, *change)
				}
			}

			prices[assetID] = price
		}
		c.JSON(200, prices)
		return
	}

	// Single asset case
	price, err := a.priceService.GetLastPrice(c.Request.Context(), asset)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch last price"})
		return
	}

	// Calculate price changes for each period
	price.PriceChanges = make([]models.PriceChange, 0, len(periodDurations))
	for period, duration := range periodDurations {
		historicalPrice, err := a.priceService.GetHistoricalPrice(c.Request.Context(), asset, duration)
		if err != nil {
			zap.L().Error("Failed to fetch historical price",
				zap.String("asset", asset),
				zap.String("period", period),
				zap.Error(err))
			continue
		}

		if change := models.CalculatePriceChange(price, historicalPrice, period); change != nil {
			price.PriceChanges = append(price.PriceChanges, *change)
		}
	}

	c.JSON(200, price)
}

// @Summary Stream price updates
// @Description Server-Sent Events stream of price updates
// @Tags prices
// @Produce text/event-stream
// @Success 200 {string} models.Issuance "SSE stream"
// @Router /prices/stream [get]
// func (a *API) handlePriceStream(c *gin.Context) {
// 	c.Writer.Header().Set("Content-Type", "text/event-stream")
// 	c.Writer.Header().Set("Cache-Control", "no-cache")
// 	c.Writer.Header().Set("Connection", "keep-alive")

// 	ctx := c.Request.Context()
// 	c.Stream(func(w io.Writer) bool {
// 		select {
// 		case <-ctx.Done():
// 			return false
// 		case price := <-a.priceCh:
// 			data, err := json.Marshal(price)
// 			if err != nil {
// 				zap.L().Error("Failed to marshal price", zap.Error(err))
// 				return true
// 			}
// 			logging.Logger.Info("Sending price update", zap.String("price", string(data)))
// 			c.SSEvent("price", data)
// 			return true
// 		}
// 	})
// }

func (a *API) handleIssuances(c *gin.Context) {
	var issuance models.Issuance
	if err := c.ShouldBindJSON(&issuance); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	if err := a.issuanceService.SaveIssuance(c.Request.Context(), issuance); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save issuance"})
		return
	}
	c.JSON(201, issuance)
}

// @Summary Get issuance details
// @Description Returns details of a specific issuance
// @Tags issuances
// @Accept json
// @Produce json
// @Param id path string true "Issuance ID"
// @Success 200 {object} models.Issuance
// @Router /issuances/{id} [get]
func (a *API) handleIssuance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Issuance ID required"})
		return
	}
	issuance, err := a.issuanceService.GetIssuance(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get issuance"})
		return
	}
	c.JSON(200, issuance)
}

// @Summary Get available assets
// @Description Returns list of all available assets
// @Tags assets
// @Produce json
// @Success 200 {array} models.AssetData
// @Router /assets [get]
func (a *API) handleAssets(c *gin.Context) {
	assetData := make([]models.AssetData, len(a.cfg.Assets))
	for i, asset := range a.cfg.Assets {
		assetData[i] = models.AssetData{
			AssetID: utils.GenerateIDForAsset(asset.InternalAssetIdentity),
			Asset:   asset.Name,
		}
	}
	c.JSON(200, assetData)
}

// @Summary Get price audit
// @Description Returns audit information for a specific price
// @Tags prices
// @Accept json
// @Produce json
// @Param id path string true "Price ID"
// @Success 200 {object} models.PriceAudit
// @Router /prices/{id}/audit [get]
func (a *API) handleAuditPrice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Price ID required"})
		return
	}
	priceAudit, err := a.priceService.AuditPrice(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to audit price, %v", err)})
		return
	}

	c.JSON(200, priceAudit)
}
