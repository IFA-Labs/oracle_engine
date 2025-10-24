package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"oracle_engine/internal/logging"
	"time"

	"go.uber.org/zap"
)

// ExchangeRateService handles fetching exchange rates from IFA Labs API
type ExchangeRateService struct {
	apiKey string
	apiURL string
	client *http.Client
}

// NewExchangeRateService creates a new exchange rate service
func NewExchangeRateService(apiKey, apiURL string) *ExchangeRateService {
	return &ExchangeRateService{
		apiKey: apiKey,
		apiURL: apiURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ExchangeRateResponse represents the response from IFA Labs API
type ExchangeRateResponse struct {
	AssetID   string  `json:"assetID"`
	Value     float64 `json:"value"`
	Expo      int8    `json:"expo"`
	Timestamp string  `json:"timestamp"`
	Source    string  `json:"source"`
}

// GetUSDToNGNRate fetches the current USD to NGN exchange rate from IFA Labs API
func (e *ExchangeRateService) GetUSDToNGNRate(ctx context.Context) (float64, error) {
	if e.apiKey == "" {
		logging.Logger.Warn("IFA Labs API key not configured, using fallback rate")
		return 1650.0, nil
	}

	// Construct the API URL for getting the last price for USD/NGN
	url := fmt.Sprintf("%s/api/prices/last?asset=NGN&changes=1h", e.apiURL)
	logging.Logger.Info("Fetching USD to NGN rate from IFA Labs API", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key header (assuming IFA Labs uses Authorization header)
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		logging.Logger.Warn("Failed to make request to IFA Labs API, using fallback rate",
			zap.Error(err))
		return 1650.0, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Logger.Warn("IFA Labs API returned non-OK status, using fallback rate",
			zap.Int("status_code", resp.StatusCode))
		return 1650.0, nil
	}

	var response map[string]ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logging.Logger.Warn("Failed to decode IFA Labs API response, using fallback rate",
			zap.Error(err))
		return 1650.0, nil
	}

	// Find the NGN price data
	ngnData, exists := response["NGN"]
	if !exists {
		logging.Logger.Warn("NGN price data not found in IFA Labs API response, using fallback rate")
		return 1650.0, nil
	}

	// Calculate the actual rate from the normalized value
	// The value is normalized with expo, so we need to convert it back
	rate := ngnData.Value * math.Pow(2, float64(ngnData.Expo))

	logging.Logger.Info("Successfully fetched USD to NGN rate",
		zap.Float64("rate", rate),
		zap.String("source", ngnData.Source),
		zap.String("timestamp", ngnData.Timestamp))

	return rate, nil
}

// GetExchangeRate fetches exchange rate for any currency pair
func (e *ExchangeRateService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	if e.apiKey == "" {
		logging.Logger.Warn("IFA Labs API key not configured, using fallback rate")
		return 1650.0, nil
	}

	if fromCurrency == "USD" && toCurrency == "NGN" {
		return e.GetUSDToNGNRate(ctx)
	}

	logging.Logger.Warn("Unsupported currency pair, using fallback rate",
		zap.String("from", fromCurrency),
		zap.String("to", toCurrency))
	return 1650.0, nil
}

// GetCachedRate returns a cached rate with fallback to API
func (e *ExchangeRateService) GetCachedRate(ctx context.Context) (float64, error) {
	// For now, we'll always fetch fresh data
	// In production, you might want to implement caching with Redis or similar
	return e.GetUSDToNGNRate(ctx)
}
