package pyth

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"

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
	id    string
	price struct {
		Price       string `json:"price"`
		PublishTime string `json:"publish_time"`
	}
}

func (p *PythFeed) FetchPrice(ctx context.Context) (*models.Price, error) {
	baseURL := "https://hermes.pyth.network/v2/updates/price/latest"
	params := url.Values{}
	params.Add("ids[]", "elem")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	response, err := http.Get(fullURL)
	if err != nil {
		logging.Logger.Error("Couldn't fetch data")
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logging.Logger.Error("Pyth error", zap.Any("err", responseData))
	}

	fmt.Printf("asset %v", string(responseData))

	// Pyth api call
	return &models.Price{
		Value:     50000.0, // Dummy
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
