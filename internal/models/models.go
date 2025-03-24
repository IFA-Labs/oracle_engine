package models

import (
	"time"
	"oracle_engine/internal/config"
)

type Price struct {
	Asset     string    `json:"asset"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type AssetFeed struct {
	feed config.FeedConfig
	asset string
	
}
