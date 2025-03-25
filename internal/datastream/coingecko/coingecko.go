package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"
)

type CoingeckoFeed struct {
	interval time.Duration
	assetID  string
}

func New() *CoingeckoFeed {
	return &CoingeckoFeed{}
}

type CoingeckoResponse struct {
}

func (p *CoingeckoFeed) FetchPrice(ctx context.Context, assetID string) (*models.Price, error) {

	baseURL := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/token_price/%v", assetID)
	fullURL := fmt.Sprintf("%s", baseURL)
	response, err := http.Get(fullURL)
	if err != nil {
		logging.Logger.Error("Couldn't fetch data")
		return nil, err
	}

	responseData, _ := io.ReadAll(response.Body)

	var coingeckoResponse CoingeckoResponse
	err = json.Unmarshal(responseData, &coingeckoResponse)
	if err != nil {
		errMsg := fmt.Errorf("error unmarshaling %w", err)
		fmt.Printf("%v", errMsg)
		return nil, err
	}

	// Pyth api call
	return &models.Price{
		Value:     3.0,
		Timestamp: time.Now(),
		Source:    p.Name(),
	}, nil
}

func (p *CoingeckoFeed) Name() string {
	return "pyth"
}

func (p *CoingeckoFeed) Interval() time.Duration {
	return p.interval // Default, overridden by config.yaml
}

func (p *CoingeckoFeed) AssetID() string {
	return p.assetID
}
