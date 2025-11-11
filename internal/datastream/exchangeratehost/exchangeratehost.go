package exchangeratehost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"strings"
	"time"

	"go.uber.org/zap"
)

type ExchangeRateHostFeed struct {
	interval time.Duration
	assetID  string
	client   *http.Client
}

func New(cfg *config.Config) *ExchangeRateHostFeed {
	return &ExchangeRateHostFeed{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type exchangeRateHostResponse struct {
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

func normalizeCurrency(code string) string {
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

func (e *ExchangeRateHostFeed) FetchPrice(ctx context.Context, assetID string, quoteAssetID string, internalAssetID string) (*models.Price, error) {
	baseCurrency := normalizeCurrency(assetID)

	targetCurrency := quoteAssetID
	if targetCurrency == "" {
		targetCurrency = "USD"
	}
	targetCurrency = normalizeCurrency(targetCurrency)

	params := url.Values{}
	params.Add("from", baseCurrency)
	params.Add("to", targetCurrency)
	params.Add("amount", "1")

	baseURL := "https://api.exchangerate.host/convert"
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	logging.Logger.Info("Fetching ExchangeRateHost rate", zap.String("url", reqURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		logging.Logger.Error("ExchangeRateHost: failed to create request", zap.Error(err))
		return nil, err
	}

	res, err := e.client.Do(req)
	if err != nil {
		logging.Logger.Error("ExchangeRateHost: request failed", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", res.StatusCode)
		logging.Logger.Error("ExchangeRateHost: non-200 response", zap.Error(err))
		return nil, err
	}

	var apiResponse exchangeRateHostResponse
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		logging.Logger.Error("ExchangeRateHost: failed to decode response", zap.Error(err))
		return nil, err
	}

	if !apiResponse.Success {
		err := fmt.Errorf("api returned success=false")
		logging.Logger.Error("ExchangeRateHost: API error", zap.Error(err))
		return nil, err
	}

	price := apiResponse.Result

	var timestamp time.Time
	if apiResponse.Date != "" {
		parsed, err := time.Parse("2006-01-02", apiResponse.Date)
		if err == nil {
			timestamp = parsed
		}
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return &models.Price{
		Value:                 price,
		Asset:                 assetID,
		ReqURL:                reqURL,
		Expo:                  int8(0),
		Timestamp:             timestamp,
		Source:                e.Name(),
		InternalAssetIdentity: internalAssetID,
	}, nil
}

func (e *ExchangeRateHostFeed) Name() string {
	return "exchangeratehost"
}

func (e *ExchangeRateHostFeed) Interval() time.Duration {
	return e.interval
}

func (e *ExchangeRateHostFeed) AssetID() string {
	return e.assetID
}

