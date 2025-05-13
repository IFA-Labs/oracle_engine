package api

import (
	"encoding/json"
	"fmt"
	"io"
	"oracle_engine/internal/config"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/services"
	"oracle_engine/internal/utils"

	"go.uber.org/zap"

	_ "oracle_engine/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Oracle Engine API
// @version 1.0
// @description API for accessing oracle price data and issuances also to audit prices
// @host localhost:8000
// @host http://146.190.186.116:8000
// @BasePath /api
// @contact.name   Grammyboy
// @contact.url    ifa-labs
// @contact.email  support@ifa-labs.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
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

func (a *API) RegisterRoutes(router *gin.Engine) {
	// Price endpoints
	router.GET("/api/prices/last", a.handleLastPrice)
	router.GET("/api/prices/stream", a.handlePriceStream)

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
// @Success 200 {object} map[string]float64
// @Router /prices/last [get]
func (a *API) handleLastPrice(c *gin.Context) {
	asset := c.Query("asset")
	if asset == "" {
		// Return all assets' last prices
		prices := make(map[string]float64)
		for _, assetConfig := range a.cfg.Assets {
			price, err := a.priceService.GetLastPrice(c.Request.Context(), utils.GenerateIDForAsset(assetConfig.InternalAssetIdentity))
			if err != nil {
				zap.L().Error("Failed to fetch last price", zap.String("asset", assetConfig.Name), zap.Error(err))
				continue
			}
			prices[utils.GenerateIDForAsset(assetConfig.InternalAssetIdentity)] = price.Number()
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
	c.JSON(200, map[string]float64{asset: price.Number()})
}

// @Summary Stream price updates
// @Description Server-Sent Events stream of price updates
// @Tags prices
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Router /prices/stream [get]
func (a *API) handlePriceStream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ctx := c.Request.Context()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case price := <-a.priceCh:
			data, err := json.Marshal(price)
			if err != nil {
				zap.L().Error("Failed to marshal price", zap.Error(err))
				return true
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		}
	})
}

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
