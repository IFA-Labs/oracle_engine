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
	Data *MonieratePrice `json:"response"`
}

type MonieratePrice struct {
	Rate       float64 `json:"status"`
	Conversion float64 `json:"message"`
}

func (p *MonierateFeed) FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error) {

	url := "https://api.monierate.com/core/rates/convert.json"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "from": "USD",
    "to": "%s",
    "amount": 1
	}`, assetID))
	logging.Logger.Info("Monierate payload", zap.String("payload", assetID))

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

	responseData, _ := io.ReadAll(res.Body)

	var monierateResponse MonierateResponse
	err = json.Unmarshal(responseData, &monierateResponse)
	if err != nil {
		errMsg := fmt.Errorf("error unmarshaling %w", err)
		fmt.Printf("%v", errMsg)
		return nil, err
	}

	usdToAsset := monierateResponse.Data.Conversion
	logging.Logger.Info("Monierate response", zap.Float64("usdToAsset", usdToAsset))
	priceF32 := 1 / usdToAsset

	// Pyth api call
	return &models.Price{
		Value:                 priceF32,
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
