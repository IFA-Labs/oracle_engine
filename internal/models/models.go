package models

import (
	"math"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/utils"
	"time"

	"go.uber.org/zap"
)

const TargetExpo int = 18

type Price struct {
	ID                    string    `json:"id"`
	Asset                 string    `json:"asset"`
	InternalAssetIdentity string    `json:"internal_asset_identity"`
	Source                string    `json:"source"`
	ReqURL                string    `json:"req_url"`
	Value                 float64   `json:"value"`
	Expo                  int8      `json:"expo"`
	Timestamp             time.Time `json:"timestamp"`
}

type AssetFeed struct {
	feed  config.FeedConfig
	asset string
}

type UnifiedPrice struct {
	ID      string `json:"id"`
	AssetID string `json:"assetID"`
	// Cant use in64 due to overflow
	Value     float64   `json:"value"`
	Expo      int8      `json:"expo"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	ReqHash   string    `json:"req_hash"`
	// this is req url but not for aggr price
	ReqURL string `json:"req_url"`
	// is aggregated
	IsAggr            bool     `json:"is_aggr"`
	ConnectedPriceIDs []string `json:"connected_price_ids"`
}

func (p Price) ToUnified() UnifiedPrice {
	// Calculate Number and normalize
	num := p.Number()
	negativeExpo := -1 * TargetExpo
	normalized := num * math.Pow10(TargetExpo)

	return UnifiedPrice{
		ID:        p.ID,
		AssetID:   p.InternalAssetIdentity,
		IsAggr:    false,
		Value:     float64(normalized),
		Expo:      int8(negativeExpo),
		Timestamp: p.Timestamp,
		Source:    p.Source,
		ReqHash:   utils.HashWithSource(p.Source),
		ReqURL:    p.ReqURL,
	}
}

func (up Price) Number() float64 {
	// Calculate raw value (Value * 10^Expo)
	rawValue := float64(up.Value) * math.Pow10(int(up.Expo))
	logging.Logger.Warn("Here",
		zap.Any("val", up.Value),
		zap.Any("exp", up.Expo),
		zap.Any("asset", up.Asset),
	)

	return rawValue
}

// Will deprecate one of these as values will be normalized soon
func (up UnifiedPrice) Number() float64 {
	// Step 1: Calculate raw value (Value * 10^Expo)
	rawValue := float64(up.Value) * math.Pow10(int(up.Expo))

	return rawValue
}

// func (up UnifiedPrice) Normalize() int64 {
// 	number := up.Number()
// 	targetExpo := 18

// 	return int64(number * math.Pow10(targetExpo))
// }

type IssuanceState int

const (
	Denied IssuanceState = iota
	Approved
	Confirmed
)

type Issuance struct {
	ID             string        `json:"issuance_id"`
	State          IssuanceState `json:"issuance_state"`
	IssuerAddress  string        `json:"issuer_address"`
	RoundID        uint64        `json:"round_id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Price          UnifiedPrice  `json:"price"`
	PriceValue     float64       `json:"price_value"` // Normalized price value with 5 decimal places
	PriceAssetID   string        `json:"price_asset_id"`
	PriceSource    string        `json:"price_source"`
	PriceTimestamp time.Time     `json:"price_timestamp"`
	Metadata       interface{}   `json:"metadata"`
}

type PriceAudit struct {
	PriceID         string       `json:"price_id"`
	AssetID         string       `json:"asset_id"`
	AggregatedPrice UnifiedPrice `json:"aggregated_price"`
	RawPrices       []Price      `json:"raw_prices"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type AssetData struct {
	AssetID string `json:"asset_id"`
	Asset    string `json:"asset"`
}
