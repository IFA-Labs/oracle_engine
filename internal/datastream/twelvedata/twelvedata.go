package twelvedata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"

	"go.uber.org/zap"
)

type TwelveDataFeed struct {
	interval time.Duration
	assetID  string
	apiKey   string
}

func New(cfg *config.Config) *TwelveDataFeed {
	return &TwelveDataFeed{
		apiKey: cfg.ApiKeys["twelvedata"],
	}
}

type TwelveDataResponse struct {
	Symbol    string  `json:"symbol"`
	Rate      float64 `json:"rate"`
	Timestamp int64   `json:"timestamp"`
}

func (t *TwelveDataFeed) FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error) {
	// Use the provided assetID directly for the API call
	baseURL := "https://api.twelvedata.com/exchange_rate"
	params := url.Values{}
	params.Add("symbol", assetID)
	params.Add("apikey", t.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	logging.Logger.Info("Fetching TwelveData", zap.String("url", fullURL))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
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

	logging.Logger.Info("TwelveData response", zap.String("response", string(body)))

	var twelveDataResponse TwelveDataResponse
	err = json.Unmarshal(body, &twelveDataResponse)
	if err != nil {
		logging.Logger.Error("Failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	// The rate gives us the exchange rate for the specified asset pair
	// This is exactly what we want to store - the price of base currency in quote currency
	rate := twelveDataResponse.Rate

	logging.Logger.Info("TwelveData conversion",
		zap.Float64("rate", rate),
		zap.String("symbol", twelveDataResponse.Symbol),
		zap.String("description", fmt.Sprintf("Rate for %s", twelveDataResponse.Symbol)))

	// Convert timestamp from Unix to time.Time
	timestamp := time.Unix(twelveDataResponse.Timestamp, 0)

	return &models.Price{
		Asset:                 assetID,
		Value:                 rate,
		Expo:                  int8(0),
		Timestamp:             timestamp,
		Source:                t.Name(),
		InternalAssetIdentity: internalAssetId,
	}, nil
}

func (t *TwelveDataFeed) Name() string {
	return "twelvedata"
}

func (t *TwelveDataFeed) Interval() time.Duration {
	return t.interval // Default, overridden by config.yaml
}

func (t *TwelveDataFeed) AssetID() string {
	return t.assetID
}
