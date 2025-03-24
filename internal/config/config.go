package config

import (
	"log"

	"github.com/spf13/viper"
)

type FeedConfig struct {
	Name     string `mapstructure:"name"`     // e.g., "binance"
	Interval int    `mapstructure:"interval"` // Seconds (e.g., 5)
	AssetID  string `mapstructure:"assetID"`
}

type AssetConfig struct {
	Name  string       `mapstructure:"name"`  // e.g., "BTC/USD"
	Feeds []FeedConfig `mapstructure:"feeds"` // List of feeds
}

type Config struct {
	PricePoolTTL    int           `mapstructure:"price_pool_ttl"`
	AggregatorNodes int           `mapstructure:"aggregator_nodes"`
	ConsensusThresh float64       `mapstructure:"consensus_threshold"`
	Assets          []AssetConfig `mapstructure:"assets"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetDefault("price_pool_ttl", 10)
	viper.SetDefault("aggregator_nodes", 3)
	viper.SetDefault("consensus_threshold", 0.01)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Using defaults: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Config error: %v", err)
	}
	return &cfg
}
