package api

import (
	"net/http"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/server/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ExchangeRateAPI handles exchange rate related endpoints
type ExchangeRateAPI struct {
	exchangeRateService *services.ExchangeRateService
}

// NewExchangeRateAPI creates a new exchange rate API
func NewExchangeRateAPI(exchangeRateService *services.ExchangeRateService) *ExchangeRateAPI {
	return &ExchangeRateAPI{
		exchangeRateService: exchangeRateService,
	}
}

// GetUSDToNGNRateRequest represents the request to get USD to NGN rate
type GetUSDToNGNRateRequest struct {
	// No parameters needed for now, but can be extended
}

// GetUSDToNGNRate gets the current USD to NGN exchange rate
// @Summary Get USD to NGN exchange rate
// @Description Fetches the current USD to NGN exchange rate from IFA Labs API
// @Tags exchange-rates
// @Accept json
// @Produce json
// @Success 200 {object} GetUSDToNGNRateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/exchange-rates/usd-ngn [get]
func (a *ExchangeRateAPI) GetUSDToNGNRate(c *gin.Context) {
	rate, err := a.exchangeRateService.GetUSDToNGNRate(c.Request.Context())
	if err != nil {
		logging.Logger.Error("Failed to get USD to NGN rate",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch exchange rate",
		})
		return
	}

	response := GetUSDToNGNRateResponse{
		Rate:      rate,
		From:      "USD",
		To:        "NGN",
		Timestamp: "", // Will be filled by the service if needed
		Source:    "IFA Labs API",
	}

	c.JSON(http.StatusOK, response)
}

// GetExchangeRate gets exchange rate for any supported currency pair
// @Summary Get exchange rate
// @Description Fetches exchange rate for supported currency pairs
// @Tags exchange-rates
// @Accept json
// @Produce json
// @Param from query string true "From currency (e.g., USD)"
// @Param to query string true "To currency (e.g., NGN)"
// @Success 200 {object} GetUSDToNGNRateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/exchange-rates [get]
func (a *ExchangeRateAPI) GetExchangeRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Both 'from' and 'to' parameters are required",
		})
		return
	}

	rate, err := a.exchangeRateService.GetExchangeRate(c.Request.Context(), from, to)
	if err != nil {
		logging.Logger.Error("Failed to get exchange rate",
			zap.Error(err),
			zap.String("from", from),
			zap.String("to", to))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch exchange rate",
		})
		return
	}

	response := GetUSDToNGNRateResponse{
		Rate:      rate,
		From:      from,
		To:        to,
		Timestamp: "", // Will be filled by the service if needed
		Source:    "IFA Labs API",
	}

	c.JSON(http.StatusOK, response)
}
