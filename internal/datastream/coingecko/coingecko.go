package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"

	"github.com/google/uuid"
)

type CoingeckoFeed struct {
	interval time.Duration
	assetID  string
}

func New() *CoingeckoFeed {
	return &CoingeckoFeed{}
}

type CoingeckoResponse struct {
	USD float64 `json:"usd"`
}

func (p *CoingeckoFeed) FetchPrice(ctx context.Context, assetID, internalAssetId string) (*models.Price, error) {

	baseURL := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", assetID)
	fullURL := fmt.Sprintf("%s", baseURL)
	response, err := http.Get(fullURL)
	if err != nil {
		logging.Logger.Error("Couldn't fetch data")
		return nil, err
	}
	defer response.Body.Close()

	responseData, _ := io.ReadAll(response.Body)

	var coingeckoResponse map[string]CoingeckoResponse
	err = json.Unmarshal(responseData, &coingeckoResponse)
	if err != nil {
		errMsg := fmt.Errorf("error unmarshaling %w", err)
		fmt.Printf("%v", errMsg)
		return nil, err
	}

	parsed, ok := coingeckoResponse[assetID]
	if !ok {
		return nil, fmt.Errorf("missing asset key %s in coingecko response", assetID)
	}

	// Coingecko api call
	return &models.Price{
		Value:                 parsed.USD,
		Expo:                  0,
		ID:                    uuid.NewString(),
		Timestamp:             time.Now(),
		Source:                p.Name(),
		InternalAssetIdentity: internalAssetId,
		ReqURL:                fullURL,
		Asset:                 assetID,
	}, nil
}

func (p *CoingeckoFeed) Name() string {
	return "coingecko"
}

func (p *CoingeckoFeed) Interval() time.Duration {
	return p.interval // Default, overridden by config.yaml
}

func (p *CoingeckoFeed) AssetID() string {
	return p.assetID
}
