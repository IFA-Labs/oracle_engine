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

func (p *PythFeed) FetchPrice(ctx context.Context, assetID string) (*models.Price, error) {
	baseURL := "https://hermes.pyth.network/v2/updates/price/latest"
	params := url.Values{}
	params.Add("ids[]", assetID)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
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

	priceF32, err := strconv.ParseFloat(pythResponse.Parsed[0].Price.Price, 32)
	if err != nil {
		logging.Logger.Error("Couldn't parse response")
		return nil, err
	}

	// Pyth api call
	return &models.Price{
		Value:     priceF32,
		Timestamp: time.Now(),
		Source:    p.Name(),
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
