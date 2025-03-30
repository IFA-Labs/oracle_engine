package models

import (
	"math"
	"oracle_engine/internal/config"
	"oracle_engine/internal/utils"
	"time"
)

type Price struct {
	Asset                 string    `json:"asset"`
	InternalAssetIdentity string    `json:"internal_asset_identity"`
	Value                 float64   `json:"value"`
	Expo                  int8      `json:"expo"`
	Timestamp             time.Time `json:"timestamp"`
	Source                string    `json:"source"`
}

type AssetFeed struct {
	feed  config.FeedConfig
	asset string
}

type UnifiedPrice struct {
	AssetID   string    `json:"assetID"`
	Value     int64     `json:"value"`
	Expo      int8      `json:"expo"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	ReqHash   string    `json:"req_hash"`
}

func (p Price) ToUnified() UnifiedPrice {
	return UnifiedPrice{
		AssetID:   utils.GenerateIDForAsset(p.InternalAssetIdentity),
		Value:     int64(p.Value),
		Expo:      p.Expo,
		Timestamp: p.Timestamp,
		Source:    p.Source,
		ReqHash:   utils.HashWithSource(p.Source),
	}
}

func (up UnifiedPrice) Number() float64 {
	// Step 1: Calculate raw value (Value * 10^Expo)
	rawValue := float64(up.Value) * math.Pow10(int(up.Expo))

	return rawValue

	// if expoDiff >= 0 {
	// 	// Scale up to 18 decimals
	// 	return rawValue * math.Pow10(expoDiff)
	// } else {
	// 	// Scale down if Expo > 18
	// 	return rawValue / math.Pow10(-expoDiff)
	// }
}

func (up UnifiedPrice) Normalize() int64 {
	number := up.Number()
	targetExpo := 18

	return int64(number * math.Pow10(targetExpo))
}

type IssuanceState int

const (
	Denied IssuanceState = iota
	Approved
	Confirmed
)

type Issuance struct {
	ID    string        `json:"issuance_id"`
	State IssuanceState `json:"issuance_state"`
	// Todo work on metadata
	Metadata interface{}
	// price
	Price UnifiedPrice
}
