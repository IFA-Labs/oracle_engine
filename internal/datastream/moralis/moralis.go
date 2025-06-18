package moralis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"

	"go.uber.org/zap"
)

type MoralisFeed struct {
	interval time.Duration
	assetID  string
	apiKey   string
}

func New(cfg *config.Config) *MoralisFeed {
	return &MoralisFeed{
		apiKey: cfg.ApiKeys["moralis"],
	}
}

type MoralisResponse struct {
	UsdPrice     float64 `json:"usdPrice"`
	ExchangeName string  `json:"exchangeName"`
	ExchangeAddress string `json:"exchangeAddress"`
	NativePrice  struct {
		Value string `json:"value"`
		Decimals int `json:"decimals"`
		Name string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"nativePrice"`
	TokenAddress string `json:"tokenAddress"`
}

func (m *MoralisFeed) FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error) {
	// For ERC20 tokens, we need to get the USD price
	// The assetID should be the token contract address
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2.2/erc20/%s/price", assetID)
	
	logging.Logger.Info("Fetching Moralis", zap.String("url", url))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logging.Logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	// Add required headers for Moralis API
	req.Header.Add("X-API-Key", m.apiKey)
	req.Header.Add("Accept", "application/json")

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

	logging.Logger.Info("Moralis response", zap.String("response", string(body)))

	var moralisResponse MoralisResponse
	err = json.Unmarshal(body, &moralisResponse)
	if err != nil {
		logging.Logger.Error("Failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	// The usdPrice gives us the price of the token in USD
	// This is exactly what we want to store - the price of the token in USD
	tokenPrice := moralisResponse.UsdPrice

	logging.Logger.Info("Moralis conversion", 
		zap.Float64("tokenPrice", tokenPrice),
		zap.String("exchangeName", moralisResponse.ExchangeName),
		zap.String("tokenAddress", moralisResponse.TokenAddress),
		zap.String("description", "USD per 1 token"))

	return &models.Price{
		Value:                 tokenPrice,
		Expo:                  int8(0),
		Timestamp:             time.Now(),
		Source:                m.Name(),
		InternalAssetIdentity: internalAssetId,
	}, nil
}

func (m *MoralisFeed) Name() string {
	return "moralis"
}

func (m *MoralisFeed) Interval() time.Duration {
	return m.interval // Default, overridden by config.yaml
}

func (m *MoralisFeed) AssetID() string {
	return m.assetID
} 