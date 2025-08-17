package api

import (
	"fmt"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/middleware"
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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for accessing Oracle Engine endpoints.
type API struct {
	priceService     services.PriceService
	issuanceService  services.IssuanceService
	dashboardService services.DashboardService
	priceCh          chan models.Issuance
	priceStreamer    *PriceStreamer
	cfg              *config.Config
	authMiddleware   *middleware.AuthMiddleware
}

func NewAPI(priceService services.PriceService, issuanceService services.IssuanceService, dashboardService services.DashboardService, priceCh chan models.Issuance, cfg *config.Config) *API {

	priceStreamer := NewPriceStreamer(priceCh, logging.Logger)
	priceStreamer.Start()

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, dashboardService)

	return &API{
		priceService:     priceService,
		issuanceService:  issuanceService,
		dashboardService: dashboardService,
		priceCh:          priceCh,
		priceStreamer:    priceStreamer,
		cfg:              cfg,
		authMiddleware:   authMiddleware,
	}
}

func (a *API) RegisterRoutes(router *gin.Engine) {
	// Add CORS middleware for production frontend access
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // In production, replace with your frontend domain
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "X-Total-Count")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Protected price endpoints (require API key for subscription management and rate limiting)
	router.GET("/api/prices/last", a.authMiddleware.APIKeyAuth(), a.handleLastPrice)
	router.GET("/api/prices/stream", a.authMiddleware.APIKeyAuth(), a.priceStreamer.HandleStream)

	// Protected asset endpoints
	router.GET("/api/assets", a.authMiddleware.APIKeyAuth(), a.handleAssets)

	// Protected price audit endpoint
	router.GET("/api/prices/:id/audit", a.authMiddleware.APIKeyAuth(), a.handleAuditPrice)

	// Protected issuance endpoints
	router.GET("/api/issuances/:id", a.authMiddleware.APIKeyAuth(), a.handleIssuance)

	// Public authentication endpoints (no API key required)
	router.POST("/api/dashboard/signup", a.handleSignUp)
	router.POST("/api/dashboard/login", a.handleLogin)

	// Protected dashboard endpoints (require JWT authentication)
	protected := router.Group("/api/dashboard")
	protected.Use(a.authMiddleware.JWTAuth(), a.authMiddleware.ValidateProfileOwnership())
	{
		protected.GET("/:id/profile", a.handleGetProfile)
		protected.PUT("/:id/profile", a.handleUpdateProfile)
		protected.POST("/:id/api-keys", a.handleCreateAPIKey)
		protected.GET("/:id/api-keys", a.handleGetAPIKeys)
		protected.DELETE("/:id/api-keys/:key_id", a.handleDeleteAPIKey)
		protected.GET("/:id/usage", a.handleGetAPIUsage) // Get API usage statistics
		// Payment endpoints (placeholder)
		protected.POST("/:id/payment", a.handleCreatePayment)
		protected.GET("/:id/payment/history", a.handleGetPaymentHistory)
	}

	// Public subscription plan information (no auth required)
	router.GET("/api/subscription/plans", a.handleGetSubscriptionPlans)

	url := ginSwagger.URL("/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Health check endpoint
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// @Summary User Sign Up a company
// @Description Create a new company profile and user account
// @Tags dashboard
// @Accept json
// @Produce json
// @Param request body models.SignUpRequest true "Sign up request"
// @Success 201 {object} models.SignUpResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/signup [post]
func (a *API) handleSignUp(c *gin.Context) {
	var req models.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := a.dashboardService.SignUp(c.Request.Context(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to sign up: " + err.Error()})
		return
	}

	c.JSON(201, response)
}

// @Summary User Login
// @Description Login with email and password
// @Tags dashboard
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /dashboard/login [post]
func (a *API) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := a.dashboardService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(401, gin.H{"error": "Authentication failed: " + err.Error()})
		return
	}

	c.JSON(200, response)
}

// @Summary Get User Profile
// @Description Get company profile information
// @Tags dashboard
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} models.CompanyProfile
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/profile [get]
func (a *API) handleGetProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Profile ID required"})
		return
	}

	profile, err := a.dashboardService.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(200, profile)
}

// @Summary Update User Profile
// @Description Update company profile information
// @Tags dashboard
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body models.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} models.CompanyProfile
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/profile [put]
func (a *API) handleUpdateProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Profile ID required"})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	profile, err := a.dashboardService.UpdateProfile(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(200, profile)
}

// @Summary Create API Key
// @Description Create a new API key for accessing the Oracle Engine API
// @Tags dashboard
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body models.CreateAPIKeyRequest true "Create API key request"
// @Success 201 {object} models.CreateAPIKeyResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/api-keys [post]
func (a *API) handleCreateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Profile ID required"})
		return
	}

	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := a.dashboardService.CreateAPIKey(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create API key: " + err.Error()})
		return
	}

	c.JSON(201, response)
}

// @Summary Get API Keys
// @Description Get all API keys for a profile
// @Tags dashboard
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {array} models.APIKey
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/api-keys [get]
func (a *API) handleGetAPIKeys(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Profile ID required"})
		return
	}

	apiKeys, err := a.dashboardService.GetAPIKeys(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get API keys: " + err.Error()})
		return
	}

	c.JSON(200, apiKeys)
}

// @Summary Delete API Key
// @Description Delete an API key
// @Tags dashboard
// @Security BearerAuth
// @Param id path string true "Profile ID"
// @Param key_id path string true "API Key ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/api-keys/{key_id} [delete]
func (a *API) handleDeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	keyID := c.Param("key_id")

	if id == "" || keyID == "" {
		c.JSON(400, gin.H{"error": "Profile ID and Key ID required"})
		return
	}

	if err := a.dashboardService.DeleteAPIKey(c.Request.Context(), id, keyID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete API key: " + err.Error()})
		return
	}

	c.Status(204)
}

// @Summary Create Payment
// @Description Create a new payment (placeholder implementation)
// @Tags dashboard
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body models.CreatePaymentRequest true "Create payment request"
// @Success 201 {object} models.Payment
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/payment [post]
func (a *API) handleCreatePayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Profile ID required"})
		return
	}

	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	payment, err := a.dashboardService.CreatePayment(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create payment: " + err.Error()})
		return
	}

	c.JSON(201, payment)
}

// @Summary Get Payment History
// @Description Get payment history for a profile
// @Tags dashboard
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} models.PaymentHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dashboard/{id}/payment/history [get]
func (a *API) handleGetPaymentHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Profile ID required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := a.dashboardService.GetPaymentHistory(c.Request.Context(), id, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get payment history: " + err.Error()})
		return
	}

	c.JSON(200, response)
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

// @Summary Get API usage statistics
// @Description Returns API usage statistics for the authenticated user
// @Tags dashboard
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} object
// @Failure 401 {object} object
// @Security BearerAuth
// @Router /dashboard/{id}/usage [get]
func (a *API) handleGetAPIUsage(c *gin.Context) {
	profileID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get API usage records
	usage, err := a.dashboardService.GetAPIUsage(c.Request.Context(), profileID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get API usage"})
		return
	}

	c.JSON(200, gin.H{
		"usage": usage,
		"page":  page,
		"limit": limit,
	})
}

// @Summary Get subscription plans
// @Description Returns available subscription plans with pricing and limits
// @Tags subscription
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Router /subscription/plans [get]
func (a *API) handleGetSubscriptionPlans(c *gin.Context) {
	c.JSON(200, gin.H{
		"plans": a.cfg.SubscriptionPlans,
	})
}
