package monierate

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

	"github.com/google/uuid"

	"go.uber.org/zap"
)

type MonierateFeed struct {
	interval time.Duration
	assetID  string
	apiKey   string
}

func New(cfg *config.Config) *MonierateFeed {
	return &MonierateFeed{
		apiKey: cfg.ApiKeys["monierate"],
	}
}

type MonierateResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    *MonieratePrice `json:"data"` // ✅ corrected from "response"
}

type MonieratePrice struct {
	Rate       float64 `json:"rate"`
	Conversion float64 `json:"conversion"`
	Timestamp  int64   `json:"timestamp"`
}

func normalizeCurrencyCode(code string) string {
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

func (p *MonierateFeed) FetchPrice(ctx context.Context, assetID string, quoteAssetID string, internalAssetId string) (*models.Price, error) {
	// When quoteAssetID is provided, we want to convert FROM assetID TO quoteAssetID
	// For example, ETH/cNGN means: 1 ETH = X cNGN, so convert FROM ETH TO cNGN
	var fromCurrency, toCurrency string
	
	if quoteAssetID != "" {
		// Convert from base asset to quote asset (e.g., ETH to cNGN)
		fromCurrency = normalizeCurrencyCode(assetID)
		toCurrency = normalizeCurrencyCode(quoteAssetID)
	} else {
		// Default to USD if no quote asset specified
		fromCurrency = normalizeCurrencyCode("USD")
		toCurrency = normalizeCurrencyCode(assetID)
	}

	url := "https://api.monierate.com/core/rates/convert.json"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "from": "%s",
    "to": "%s",
    "amount": 1,
	"market": "parallel"
	}`, fromCurrency, toCurrency))
	logging.Logger.Info("Monierate payload",
		zap.String("from", fromCurrency),
		zap.String("to", toCurrency),
	)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("api_key", p.apiKey)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	logging.Logger.Info("Monierate response", zap.String("response", string(body)))

	var monierateResponse MonierateResponse
	err = json.Unmarshal(body, &monierateResponse)
	if err != nil {
		errMsg := fmt.Errorf("error unmarshaling %w", err)
		fmt.Printf("%v", errMsg)
		return nil, err
	}

	if monierateResponse.Data == nil {
		return nil, fmt.Errorf("monierate response missing data")
	}

	// Conversion is the amount of toCurrency you get for 1 fromCurrency
	// For ETH/cNGN: Conversion = amount of cNGN for 1 ETH
	conversion := monierateResponse.Data.Conversion
	logging.Logger.Info("Monierate response", 
		zap.Float64("conversion", conversion),
		zap.String("from", fromCurrency),
		zap.String("to", toCurrency),
	)
	priceF32 := float32(conversion)

	// Pyth api call
	return &models.Price{
		ID:                    uuid.NewString(),
		ReqURL:                url,
		Value:                 priceF32,
		Asset:                 assetID,
		Expo:                  int8(0),
		Timestamp:             time.Now(),
		Source:                p.Name(),
		InternalAssetIdentity: internalAssetId,
	}, nil
}

func (p *MonierateFeed) Name() string {
	return "monierate"
}

func (p *MonierateFeed) Interval() time.Duration {
	return p.interval // Default, overridden by config.yaml
}

func (p *MonierateFeed) AssetID() string {
	return p.assetID
}
