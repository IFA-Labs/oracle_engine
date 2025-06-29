package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type FeedConfig struct {
	Name     string `mapstructure:"name"`     // e.g., "binance"
	Interval int    `mapstructure:"interval"` // Seconds (e.g., 5)
	AssetID  string `mapstructure:"assetID"`
}

type ContractConfig struct {
	Address   string `mapstructure:"address"`
	RPC       string `mapstructure:"rpc"`
	ABI       string `mapstructure:"abi"`
	ChainID   string `mapstructure:"chainID"`
	ChainName string `mapstructure:"chainName"`
}

type AssetConfig struct {
	Name                  string       `mapstructure:"name"`                  // e.g., "BTC/USD"
	InternalAssetIdentity string       `mapstructure:"internalAssetIdentity"` // eg "0xUSDT"
	Feeds                 []FeedConfig `mapstructure:"feeds"`                 // List of feeds
}

type ApiKey map[string]string

type Config struct {
	PricePoolTTL         int              `mapstructure:"price_pool_ttl"`
	RELAY_TIME_THRESHOLD int              `mapstructure:"RELAY_TIME_THRESHOLD"`
	AggregatorNodes      int              `mapstructure:"aggregator_nodes"`
	ConsensusThresh      float64          `mapstructure:"consensus_threshold"`
	AggrDevPerc          float32          `mapstructure:"aggr_dev_perc"`
	Assets               []AssetConfig    `mapstructure:"assets"`
	ApiKeys              ApiKey           `mapstructure:"api_keys"`
	Contracts            []ContractConfig `mapstructure:"contracts"`
	PrivateKey           string           `mapstructure:"private_key"`
	DB_URL               string           `mapstructure:"DB_URL"`
	SERVER_PORT          string           `mapstructure:"server_port"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetDefault("price_pool_ttl", 10)
	viper.SetDefault("aggregator_nodes", 3)
	viper.SetDefault("consensus_threshold", 0.01)
	viper.SetDefault("api_keys", map[string]string{
		"monierate": os.Getenv("MONIERATE_API_KEY"),
		"exchangerate": os.Getenv("EXCHANGERATE_API_KEY"),
		"twelvedata": os.Getenv("TWELVEDATA_API_KEY"),
		"fixer": os.Getenv("FIXER_API_KEY"),
		"currencylayer": os.Getenv("CURRENCYLAYER_API_KEY"),
		"moralis": os.Getenv("MORALIS_API_KEY"),
	})

	_ = godotenv.Load()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Using defaults: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	if cfg.PrivateKey == "" {
		cfg.PrivateKey = os.Getenv("PRIVATE_KEY")
	}

	if cfg.DB_URL == "" {
		cfg.DB_URL = os.Getenv("DB_URL")
	}

	return &cfg
}
