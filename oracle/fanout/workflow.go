package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"oracle/contracts/evm/src/generated/ioracle"

	"github.com/ethereum/go-ethereum/common"

	pb2 "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/scheduler/cron"
	"github.com/smartcontractkit/cre-sdk-go/cre"
)

type EVMConfig struct {
	ChainName             string `json:"chainName"`
	GasLimit              uint64 `json:"gasLimit"`
	OracleContractAddress string `json:"oracleContractAddress"`
}

func (e *EVMConfig) GetChainSelector() (uint64, error) {
	return evm.ChainSelectorFromName(e.ChainName)
}

func (e *EVMConfig) NewEVMClient() (*evm.Client, error) {
	chainSelector, err := e.GetChainSelector()
	if err != nil {
		return nil, err
	}
	return &evm.Client{
		ChainSelector: chainSelector,
	}, nil
}

type Config struct {
	Schedule     string      `json:"schedule"`
	OracleApiUrl string      `json:"oracleApiUrl"`
	EVMs         []EVMConfig `json:"evms"`
}

type PriceData struct {
	ID        string      `json:"id"`
	AssetID   string      `json:"assetID"`
	Value     json.Number `json:"value"`
	Expo      int         `json:"expo"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
}

type PricesResponse map[string]PriceData

type RawPriceResponse struct {
	Body string `consensus_aggregation:"identical" json:"body"`
}

type PriceFeedData struct {
	AssetIndices [][32]byte
	Prices       []ioracle.IIfaPriceFeedPriceFeed
}

func InitWorkflow(config *Config, logger *slog.Logger, secretsProvider cre.SecretsProvider) (cre.Workflow[*Config], error) {
	cronTriggerCfg := &cron.Config{
		Schedule: config.Schedule,
	}

	workflow := cre.Workflow[*Config]{
		cre.Handler(
			cron.Trigger(cronTriggerCfg),
			onPriceFeedCronTrigger,
		),
	}

	return workflow, nil
}

func onPriceFeedCronTrigger(config *Config, runtime cre.Runtime, outputs *cron.Payload) (string, error) {
	return submitPriceFeeds(config, runtime)
}

func submitPriceFeeds(config *Config, runtime cre.Runtime) (string, error) {
	logger := runtime.Logger()
	logger.Info("fetching prices", "url", config.OracleApiUrl)

	client := &http.Client{}
	rawResp, err := http.SendRequest(config, runtime, client, fetchRawPrices, cre.ConsensusAggregationFromTags[*RawPriceResponse]()).Await()
	if err != nil {
		logger.Error("error fetching prices", "err", err)
		return "", err
	}

	priceFeedData, err := parsePriceResponse(rawResp.Body, logger)
	if err != nil {
		logger.Error("error parsing prices", "err", err)
		return "", err
	}

	logger.Info("fetched prices", "count", len(priceFeedData.AssetIndices))

	if len(priceFeedData.AssetIndices) == 0 {
		logger.Info("no prices to submit")
		return "no prices", nil
	}

	for _, evmCfg := range config.EVMs {
		if err := submitPriceFeedToChain(runtime, evmCfg, priceFeedData); err != nil {
			logger.Error("failed to submit price feed", "chain", evmCfg.ChainName, "err", err)
			return "", fmt.Errorf("failed to submit to %s: %w", evmCfg.ChainName, err)
		}
		logger.Info("submitted price feed", "chain", evmCfg.ChainName, "priceCount", len(priceFeedData.AssetIndices))
	}

	return fmt.Sprintf("submitted %d prices to %d chains", len(priceFeedData.AssetIndices), len(config.EVMs)), nil
}

func submitPriceFeedToChain(runtime cre.Runtime, evmCfg EVMConfig, priceFeedData *PriceFeedData) error {
	logger := runtime.Logger()

	evmClient, err := evmCfg.NewEVMClient()
	if err != nil {
		return fmt.Errorf("failed to create EVM client for %s: %w", evmCfg.ChainName, err)
	}

	oracleAddress := common.HexToAddress(evmCfg.OracleContractAddress)
	oracle, err := ioracle.NewIOracle(evmClient, oracleAddress, nil)
	if err != nil {
		return fmt.Errorf("failed to create oracle contract: %w", err)
	}

	input := ioracle.SubmitPriceFeedInput{
		Assetindex: priceFeedData.AssetIndices,
		Prices:     priceFeedData.Prices,
	}

	encoded, err := oracle.Codec.EncodeSubmitPriceFeedMethodCall(input)
	if err != nil {
		return fmt.Errorf("failed to encode submitPriceFeed call: %w", err)
	}

	reportPromise := runtime.GenerateReport(&pb2.ReportRequest{
		EncodedPayload: encoded,
		EncoderName:    "evm",
		SigningAlgo:    "ecdsa",
		HashingAlgo:    "keccak256",
	})

	report, err := reportPromise.Await()
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	var gasConfig *evm.GasConfig
	if evmCfg.GasLimit > 0 {
		gasConfig = &evm.GasConfig{
			GasLimit: evmCfg.GasLimit,
		}
	}

	resp, err := oracle.WriteReport(runtime, report, gasConfig).Await()
	if err != nil {
		logger.Error("WriteReport failed", "error", err, "chain", evmCfg.ChainName)
		return fmt.Errorf("failed to write report: %w", err)
	}

	logger.Info("price feed submitted", "txHash", common.BytesToHash(resp.TxHash).Hex(), "chain", evmCfg.ChainName)
	return nil
}

func fetchRawPrices(config *Config, logger *slog.Logger, sendRequester *http.SendRequester) (*RawPriceResponse, error) {
	httpResp, err := sendRequester.SendRequest(&http.Request{
		Method: "GET",
		Url:    config.OracleApiUrl + "/api/prices/last",
	}).Await()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	return &RawPriceResponse{
		Body: string(httpResp.Body),
	}, nil
}

func parsePriceResponse(body string, logger *slog.Logger) (*PriceFeedData, error) {
	var pricesResp PricesResponse
	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.UseNumber()
	if err := decoder.Decode(&pricesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prices response: %w", err)
	}

	assetIndices := make([][32]byte, 0, len(pricesResp))
	prices := make([]ioracle.IIfaPriceFeedPriceFeed, 0, len(pricesResp))

	for assetID, priceData := range pricesResp {
		assetIndex, err := hexToBytes32(assetID)
		if err != nil {
			logger.Warn("skipping invalid asset ID", "assetID", assetID, "err", err)
			continue
		}

		priceInt, err := convertPriceToInt256(priceData.Value)
		if err != nil {
			logger.Warn("skipping invalid price value", "assetID", assetID, "err", err)
			continue
		}
		decimal := int8(priceData.Expo)
		lastUpdateTime := uint64(priceData.Timestamp.Unix())

		assetIndices = append(assetIndices, assetIndex)
		prices = append(prices, ioracle.IIfaPriceFeedPriceFeed{
			Price:          priceInt,
			Decimal:        decimal,
			LastUpdateTime: lastUpdateTime,
		})

		logger.Info("prepared price", "assetID", assetID[:16]+"...", "price", priceInt.String(), "decimal", decimal, "timestamp", lastUpdateTime)
	}

	return &PriceFeedData{
		AssetIndices: assetIndices,
		Prices:       prices,
	}, nil
}

func hexToBytes32(hexStr string) ([32]byte, error) {
	var result [32]byte

	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	if len(hexStr) != 64 {
		return result, fmt.Errorf("invalid hex string length: expected 64, got %d", len(hexStr))
	}

	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return result, fmt.Errorf("failed to decode hex: %w", err)
	}

	copy(result[:], decoded)
	return result, nil
}

func convertPriceToInt256(value json.Number) (*big.Int, error) {
	priceInt := new(big.Int)
	_, ok := priceInt.SetString(string(value), 10)
	if !ok {
		priceFloat, err := value.Float64()
		if err != nil {
			return nil, fmt.Errorf("failed to parse price value: %s", value)
		}
		bigFloat := big.NewFloat(priceFloat)
		bigFloat.Int(priceInt)
	}
	return priceInt, nil
}
