package currencylayer

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

type CurrencyLayerFeed struct {
	interval time.Duration
	assetID  string
	apiKey   string
}

func New(cfg *config.Config) *CurrencyLayerFeed {
	return &CurrencyLayerFeed{
		apiKey: cfg.ApiKeys["currencylayer"],
	}
}

type CurrencyLayerResponse struct {
	Success bool `json:"success"`
	Query   struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	} `json:"query"`
	Info struct {
		Rate      float64 `json:"rate"`
		Timestamp int64   `json:"timestamp"`
	} `json:"info"`
	Historical bool    `json:"historical"`
	Date       string  `json:"date"`
	Result     float64 `json:"result"`
}

func (c *CurrencyLayerFeed) FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error) {
	// For asset/USD, we need to get asset/USD rate
	// Since CurrencyLayer API uses from/to format, we'll convert 1 asset to USD
	baseURL := "https://api.currencylayer.com/convert"
	params := url.Values{}
	params.Add("access_key", c.apiKey)
	params.Add("from", assetID)
	params.Add("to", "USD")
	params.Add("amount", "1")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	logging.Logger.Info("Fetching CurrencyLayer", zap.String("url", fullURL))

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

	logging.Logger.Info("CurrencyLayer response", zap.String("response", string(body)))

	var currencyLayerResponse CurrencyLayerResponse
	err = json.Unmarshal(body, &currencyLayerResponse)
	if err != nil {
		logging.Logger.Error("Failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	if !currencyLayerResponse.Success {
		errMsg := fmt.Errorf("API returned success=false")
		logging.Logger.Error("API error", zap.Error(errMsg))
		return nil, errMsg
	}

	// The result gives us how many USD we get for 1 asset
	// This is exactly what we want to store - the price of asset in USD
	assetToUsd := currencyLayerResponse.Result

	logging.Logger.Info("CurrencyLayer conversion",
		zap.Float64("brlToUsd", assetToUsd),
		zap.Float64("rate", currencyLayerResponse.Info.Rate),
		zap.String("from", currencyLayerResponse.Query.From),
		zap.String("to", currencyLayerResponse.Query.To),
		zap.String("description", "USD per 1 asset"))

	// Convert timestamp from Unix to time.Time if available
	var timestamp time.Time
	if currencyLayerResponse.Info.Timestamp > 0 {
		timestamp = time.Unix(currencyLayerResponse.Info.Timestamp, 0)
	} else {
		timestamp = time.Now()
	}

	return &models.Price{
		Value:                 assetToUsd,
		Expo:                  int8(0),
		Timestamp:             timestamp,
		Source:                c.Name(),
		InternalAssetIdentity: internalAssetId,
		Asset:                 assetID,
	}, nil
}

func (c *CurrencyLayerFeed) Name() string {
	return "currencylayer"
}

func (c *CurrencyLayerFeed) Interval() time.Duration {
	return c.interval // Default, overridden by config.yaml
}

func (c *CurrencyLayerFeed) AssetID() string {
	return c.assetID
}
