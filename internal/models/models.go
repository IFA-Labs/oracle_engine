package models

import "time"

type Price struct {
	Asset     string    `json:"asset"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}
