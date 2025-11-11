package exchangerate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"strings"
	"time"

	"go.uber.org/zap"
)

type ExchangeRateFeed struct {
	interval time.Duration
	assetID  string
	apiKey   string
}

func New(cfg *config.Config) *ExchangeRateFeed {
	return &ExchangeRateFeed{
		apiKey: cfg.ApiKeys["exchangerate"],
	}
}

type ExchangeRateResponse struct {
	Result             string  `json:"result"`
	Documentation      string  `json:"documentation"`
	TermsOfUse         string  `json:"terms_of_use"`
	TimeLastUpdateUnix int64   `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string  `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64   `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string  `json:"time_next_update_utc"`
	BaseCode           string  `json:"base_code"`
	TargetCode         string  `json:"target_code"`
	ConversionRate     float64 `json:"conversion_rate"`
}

func normalizeCurrency(code string) string {
	if code == "" {
		return ""
	}
	switch strings.ToUpper(code) {
	case "CNGN":
		return "NGN"
	default:
		return strings.ToUpper(code)
	}
}

func (e *ExchangeRateFeed) FetchPrice(ctx context.Context, assetID string, quoteAssetID string, internalAssetId string) (*models.Price, error) {
	targetCurrency := quoteAssetID
	if targetCurrency == "" {
		targetCurrency = "USD"
	}
	baseCurrency := normalizeCurrency(assetID)
	targetCurrency = normalizeCurrency(targetCurrency)

	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/pair/%s/%s", e.apiKey, baseCurrency, targetCurrency)

	logging.Logger.Info("Fetching ExchangeRate", zap.String("url", url))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logging.Logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		logging.Logger.Error("Failed to make request", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logging.Logger.Error("Failed to read response body", zap.Error(err))
		return nil, err
	}

	logging.Logger.Info("ExchangeRate response", zap.String("response", string(body)))

	var exchangeRateResponse ExchangeRateResponse
	err = json.Unmarshal(body, &exchangeRateResponse)
	if err != nil {
		logging.Logger.Error("Failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	if exchangeRateResponse.Result != "success" {
		errMsg := fmt.Errorf("API returned error result: %s", exchangeRateResponse.Result)
		logging.Logger.Error("API error", zap.Error(errMsg))
		return nil, errMsg
	}

	// The conversion_rate gives us <asset>/USD rate (how many USD for 1 Asset)
	// This is exactly what we want to store - the price of asset in USD
	assetToUsd := exchangeRateResponse.ConversionRate

	logging.Logger.Info("ExchangeRate conversion",
		zap.Float64("brlToUsd", assetToUsd),
		zap.String("description", "USD per 1 BRL"))

	return &models.Price{
		Value:                 assetToUsd,
		Asset:                 assetID,
		Expo:                  int8(0),
		Timestamp:             time.Now(),
		Source:                e.Name(),
		InternalAssetIdentity: internalAssetId,
	}, nil
}

func (e *ExchangeRateFeed) Name() string {
	return "exchangerate"
}

func (e *ExchangeRateFeed) Interval() time.Duration {
	return e.interval // Default, overridden by config.yaml
}

func (e *ExchangeRateFeed) AssetID() string {
	return e.assetID
}
