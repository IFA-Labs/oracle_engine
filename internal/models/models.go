package models

import (
	"math"
	"oracle_engine/internal/config"
	"time"
)

type Price struct {
	Asset     string    `json:"asset"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type AssetFeed struct {
	feed  config.FeedConfig
	asset string
}

type UnifiedPrice struct {
	AssetID   string    `json:"assetID"`
	Value     int32     `json:"value"`
	Expo      int8      `json:"expo"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	ReqHash   string    `json:"req_hash"`
}

func (up UnifiedPrice) Number() float64 {
	// Step 1: Calculate raw value (Value * 10^Expo)
	rawValue := float64(up.Value) * math.Pow10(int(up.Expo))

	// Step 2: Normalize to 18 decimal places
	targetExpo := 18
	// expoDiff := targetExpo - int(up.Expo)

	return rawValue * math.Pow10(targetExpo)

	// if expoDiff >= 0 {
	// 	// Scale up to 18 decimals
	// 	return rawValue * math.Pow10(expoDiff)
	// } else {
	// 	// Scale down if Expo > 18
	// 	return rawValue / math.Pow10(-expoDiff)
	// }
}
