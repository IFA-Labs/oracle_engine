package pyth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PythFeed struct {
	interval time.Duration
	assetID  string
}

func New() *PythFeed {
	return &PythFeed{}
}

type PythResponse struct {
	Parsed []*PythPrice `json:"parsed"`
}

type PythPrice struct {
	Id    string `json:"id"`
	Price struct {
		Price       string `json:"price"`
		PublishTime int    `json:"publish_time"`
		Exponential int    `json:"expo"`
	} `json:"price"`
}

func (p *PythFeed) FetchPrice(ctx context.Context, assetID string, internalAssetId string) (*models.Price, error) {
	baseURL := "https://hermes.pyth.network/v2/updates/price/latest"
	params := url.Values{}
	params.Add("ids[]", assetID)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	logging.Logger.Info("hel", zap.Any("url", p.assetID))
	response, err := http.Get(fullURL)
	if err != nil {
		logging.Logger.Error("Couldn't fetch data")
		return nil, err
	}

	responseData, _ := io.ReadAll(response.Body)

	var pythResponse PythResponse
	err = json.Unmarshal(responseData, &pythResponse)
	if err != nil {
		errMsg := fmt.Errorf("error unmarshaling %w", err)
		fmt.Printf("%v", errMsg)
		return nil, err
	}

	priceElem := pythResponse.Parsed[0].Price
	priceF32, err := strconv.ParseFloat(priceElem.Price, 32)
	if err != nil {
		logging.Logger.Error("Couldn't parse response")
		return nil, err
	}

	// Pyth api call
	return &models.Price{
		Value:                 priceF32,
		Expo:                  int8(priceElem.Exponential),
		Timestamp:             time.Now(),
		Source:                p.Name(),
		InternalAssetIdentity: internalAssetId,
		Asset:                 assetID,
		ID:                    uuid.NewString(),
		ReqURL:                fullURL,
	}, nil
}

func (p *PythFeed) Name() string {
	return "pyth"
}

func (p *PythFeed) Interval() time.Duration {
	return p.interval // Default, overridden by config.yaml
}

func (p *PythFeed) AssetID() string {
	return p.assetID
}
