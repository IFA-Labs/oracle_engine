package api

import (
	"net/http"
	"time"

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

// GetExchangeRateResponse represents the response for exchange rate queries
type GetExchangeRateResponse struct {
	Rate      float64 `json:"rate" example:"1560.35"`
	From      string  `json:"from" example:"USD"`
	To        string  `json:"to" example:"NGN"`
	Timestamp int64   `json:"timestamp" example:"1700000000"`
	Source    string  `json:"source" example:"IFA Labs API"`
}

// GetErrorResponse represents an error message
type GetErrorResponse struct {
	Error string `json:"error" example:"Failed to fetch exchange rate"`
}

// ----------------------------------------------------------------------
// Get USD → NGN Rate
// ----------------------------------------------------------------------

// GetUSDToNGNRate retrieves the current USD to NGN exchange rate
// @Summary Get USD → NGN exchange rate
// @Description Fetches the current USD → NGN exchange rate from the IFA Labs API.
// @Tags Exchange Rates
// @Produce json
// @Success 200 {object} GetExchangeRateResponse
// @Failure 500 {object} GetErrorResponse
// @Router /api/exchange-rates/usd-ngn [get]
func (a *ExchangeRateAPI) GetUSDToNGNRate(c *gin.Context) {
	rate, err := a.exchangeRateService.GetUSDToNGNRate(c.Request.Context())
	if err != nil {
		logging.Logger.Error("Failed to fetch USD to NGN rate", zap.Error(err))

		c.JSON(http.StatusInternalServerError, GetErrorResponse{
			Error: "Failed to fetch exchange rate",
		})
		return
	}

	response := GetExchangeRateResponse{
		Rate:      rate,
		From:      "USD",
		To:        "NGN",
		Timestamp: time.Now().Unix(),
		Source:    "IFA Labs API",
	}

	c.JSON(http.StatusOK, response)
}

// ----------------------------------------------------------------------
// Generic Exchange Rate
// ----------------------------------------------------------------------

// GetExchangeRate retrieves exchange rates for any supported currency pair
// @Summary Get exchange rate for any currency pair
// @Description Fetches the exchange rate for any supported currency pair (e.g., USD → NGN, GBP → USD).
// @Tags Exchange Rates
// @Produce json
// @Param from query string true "Base currency" example(USD)
// @Param to query string true "Target currency" example(NGN)
// @Success 200 {object} GetExchangeRateResponse
// @Failure 400 {object} GetErrorResponse
// @Failure 500 {object} GetErrorResponse
// @Router /api/exchange-rates [get]
func (a *ExchangeRateAPI) GetExchangeRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, GetErrorResponse{
			Error: "Both 'from' and 'to' parameters are required",
		})
		return
	}

	rate, err := a.exchangeRateService.GetExchangeRate(c.Request.Context(), from, to)
	if err != nil {
		logging.Logger.Error("Failed to get exchange rate",
			zap.Error(err),
			zap.String("from", from),
			zap.String("to", to),
		)

		c.JSON(http.StatusInternalServerError, GetErrorResponse{
			Error: "Failed to fetch exchange rate",
		})
		return
	}

	response := GetExchangeRateResponse{
		Rate:      rate,
		From:      from,
		To:        to,
		Timestamp: time.Now().Unix(),
		Source:    "IFA Labs API",
	}

	c.JSON(http.StatusOK, response)
}
