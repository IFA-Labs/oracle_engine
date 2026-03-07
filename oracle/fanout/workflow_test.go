package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	evmmock "github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm/mock"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http"
	httpmock "github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http/mock"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/scheduler/cron"
	"github.com/smartcontractkit/cre-sdk-go/cre/testutils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"oracle/contracts/evm/src/generated/ioracle"
)

var anyExecutionTime = time.Unix(1752514917, 0)

func TestInitWorkflow(t *testing.T) {
	config := makeTestConfig(t)
	runtime := testutils.NewRuntime(t, testutils.Secrets{})

	workflow, err := InitWorkflow(config, runtime.Logger(), nil)
	require.NoError(t, err)

	require.Len(t, workflow, 1)
	require.Equal(t, cron.Trigger(&cron.Config{}).CapabilityID(), workflow[0].CapabilityID())
}

func TestOnCronTrigger(t *testing.T) {
	config := makeTestConfig(t)
	runtime := testutils.NewRuntime(t, testutils.Secrets{
		"": {},
	})

	httpMock, err := httpmock.NewClientCapability(t)
	require.NoError(t, err)
	httpMock.SendRequest = func(ctx context.Context, input *http.Request) (*http.Response, error) {
		pricesResponse := `{
			"6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d": {
				"id": "e7d07dd9-1caf-483b-9797-08f4dba40dff",
				"assetID": "6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d",
				"value": 999954818031027100,
				"expo": -18,
				"timestamp": "2026-03-07T08:53:13.041313Z",
				"source": "ifa_labs"
			},
			"8c3fb07cab369fe230ca4e45d095f796c4c1a30131f1799766d4fec5ee1325c0": {
				"id": "e7bc6f4f-b7c9-4dfd-ab8e-b6424aac77b5",
				"assetID": "8c3fb07cab369fe230ca4e45d095f796c4c1a30131f1799766d4fec5ee1325c0",
				"value": 1982146157076351500000,
				"expo": -18,
				"timestamp": "2026-03-07T08:54:07.969995Z",
				"source": "ifa_labs"
			}
		}`
		return &http.Response{Body: []byte(pricesResponse)}, nil
	}

	chainSelector, err := config.EVMs[0].GetChainSelector()
	require.NoError(t, err)
	evmMock, err := evmmock.NewClientCapability(chainSelector, t)
	require.NoError(t, err)

	evmCfg := config.EVMs[0]
	_ = ioracle.NewIOracleMock(
		common.HexToAddress(evmCfg.OracleContractAddress),
		evmMock,
	)

	evmMock.WriteReport = func(ctx context.Context, input *evm.WriteReportRequest) (*evm.WriteReportReply, error) {
		return &evm.WriteReportReply{
			TxHash: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef").Bytes(),
		}, nil
	}

	result, err := onPriceFeedCronTrigger(config, runtime, &cron.Payload{
		ScheduledExecutionTime: timestamppb.New(anyExecutionTime),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Contains(t, result, "submitted")
	require.Contains(t, result, "prices")

	logs := runtime.GetLogs()
	assertLogContains(t, logs, `msg="fetching prices"`)
	assertLogContains(t, logs, `msg="fetched prices"`)
}

func TestFetchPrices(t *testing.T) {
	pricesResponse := `{
		"6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d": {
			"id": "e7d07dd9-1caf-483b-9797-08f4dba40dff",
			"assetID": "6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d",
			"value": 999954818031027100,
			"expo": -18,
			"timestamp": "2026-03-07T08:53:13.041313Z",
			"source": "ifa_labs"
		}
	}`

	var pricesResp PricesResponse
	err := json.Unmarshal([]byte(pricesResponse), &pricesResp)
	require.NoError(t, err)
	require.Len(t, pricesResp, 1)

	assetID := "6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d"
	price, ok := pricesResp[assetID]
	require.True(t, ok)
	require.Equal(t, assetID, price.AssetID)
	require.Equal(t, float64(999954818031027100), price.Value)
	require.Equal(t, -18, price.Expo)
}

func TestHexToBytes32(t *testing.T) {
	hexStr := "6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d"
	result, err := hexToBytes32(hexStr)
	require.NoError(t, err)
	require.Len(t, result, 32)

	hexStrWithPrefix := "0x6ca0cef6107263f3b09a51448617b659278cff744f0e702c24a2f88c91e65a0d"
	result2, err := hexToBytes32(hexStrWithPrefix)
	require.NoError(t, err)
	require.Equal(t, result, result2)

	_, err = hexToBytes32("invalid")
	require.Error(t, err)
}

func TestConvertPriceToInt256(t *testing.T) {
	priceInt, err := convertPriceToInt256(json.Number("999954818031027100"))
	require.NoError(t, err)
	require.NotNil(t, priceInt)
	require.Equal(t, "999954818031027100", priceInt.String())
}

//go:embed config.production.json
var configJson []byte

func makeTestConfig(t *testing.T) *Config {
	config := &Config{}
	require.NoError(t, json.Unmarshal(configJson, config))
	return config
}

func assertLogContains(t *testing.T, logs [][]byte, substr string) {
	for _, line := range logs {
		if strings.Contains(string(line), substr) {
			return
		}
	}
	t.Fatalf("Expected logs to contain substring %q, but it was not found in logs:\n%s",
		substr, strings.Join(func() []string {
			var logStrings []string
			for _, log := range logs {
				logStrings = append(logStrings, string(log))
			}
			return logStrings
		}(), "\n"))
}
