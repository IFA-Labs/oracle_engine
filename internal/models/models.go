package models

import (
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
