package fixer

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

type FixerFeed struct {
	interval time.Duration
	assetID  string
	apiKey   string
}

func New(cfg *config.Config) *FixerFeed {
	return &FixerFeed{
		apiKey: cfg.ApiKeys["fixer"],
	}
}

type FixerResponse struct {
	Success bool `json:"success"`
	Query   struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	} `json:"query"`
	Info struct {
		Rate float64 `json:"rate"`
	} `json:"info"`
	Historical bool    `json:"historical"`
	Date       string  `json:"date"`
	Result     float64 `json:"result"`
}

func (f *FixerFeed) FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error) {
	// Parse assetID to get base and quote currencies
	// Expected format: "BASE/QUOTE" (e.g., "BRL/USD")
	baseURL := "https://data.fixer.io/api/convert"
	params := url.Values{}
	params.Add("access_key", f.apiKey)
	params.Add("from", assetID)
	params.Add("to", "USD")
	params.Add("amount", "1")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	logging.Logger.Info("Fetching Fixer", zap.String("url", fullURL))

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

	logging.Logger.Info("Fixer response", zap.String("response", string(body)))

	var fixerResponse FixerResponse
	err = json.Unmarshal(body, &fixerResponse)
	if err != nil {
		logging.Logger.Error("Failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	if !fixerResponse.Success {
		errMsg := fmt.Errorf("API returned success=false")
		logging.Logger.Error("API error", zap.Error(errMsg))
		return nil, errMsg
	}

	// The result gives us how many quote currency we get for 1 base currency
	// This is exactly what we want to store - the price of base currency in quote currency
	rate := fixerResponse.Result

	logging.Logger.Info("Fixer conversion",
		zap.Float64("rate", rate),
		zap.Float64("apiRate", fixerResponse.Info.Rate),
		zap.String("from", fixerResponse.Query.From),
		zap.String("to", fixerResponse.Query.To),
		zap.String("description", fmt.Sprintf("%s per 1 %s", fixerResponse.Query.To, fixerResponse.Query.From)))

	return &models.Price{
		Asset:                 assetID,
		Value:                 rate,
		Expo:                  int8(0),
		Timestamp:             time.Now(),
		Source:                f.Name(),
		InternalAssetIdentity: internalAssetId,
	}, nil
}

func (f *FixerFeed) Name() string {
	return "fixer"
}

func (f *FixerFeed) Interval() time.Duration {
	return f.interval // Default, overridden by config.yaml
}

func (f *FixerFeed) AssetID() string {
	return f.assetID
}
