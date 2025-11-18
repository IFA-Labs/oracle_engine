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
	QuoteAssetID string `mapstructure:"quoteAssetID"`
}

type ContractConfig struct {
	Address   string `mapstructure:"address"`
	RPC       string `mapstructure:"rpc"`
	ABI       string `mapstructure:"abi"`
	ChainID   string `mapstructure:"chainID"`
	ChainName string `mapstructure:"chainName"`
}

type AssetSetting struct {
	// ttl in seconds
	TTL int `mapstructure:"ttl"` // Time to live for price pool
	// percentage deviation for consensus in perc eg 0.01
	DevPerc float32 `mapstructure:"dev_perc"` // Deviation percentage for consensus
}

type AssetConfig struct {
	Name                  string       `mapstructure:"name"`                  // e.g., "BTC/USD"
	InternalAssetIdentity string       `mapstructure:"internalAssetIdentity"` // eg "0xUSDT"
	Feeds                 []FeedConfig `mapstructure:"feeds"`                 // List of feeds
	// Settings
	Settings AssetSetting `mapstructure:"settings"` // Settings for the asset
}

var DefaultAssetSetting = AssetSetting{
	TTL:     10,   // Default TTL in seconds
	DevPerc: 0.01, // Default deviation percentage for consensus
}

type ApiKey map[string]string

type SubscriptionPlan struct {
	Name                 string  `mapstructure:"name"`
	Price                float64 `mapstructure:"price"`                    // Monthly price in USD
	SubscriptionDuration int     `mapstructure:"subscription_duration"`    // Duration in days (30 = monthly, 365 = yearly, 0 = lifetime)
	APIRequests          int64   `mapstructure:"api_requests"`             // Monthly API request limit (0 = unlimited)
	RateLimitPerHour     int     `mapstructure:"rate_limit_per_hour"`      // Requests allowed per hour (0 = unlimited)
	RateLimitPerDay      int     `mapstructure:"rate_limit_per_day"`       // Requests allowed per day (0 = unlimited)
	DataAccess           string  `mapstructure:"data_access"`              // Description of data access level
	CustomPairs          int     `mapstructure:"custom_pairs"`             // Number of custom pairs allowed
	RequestCost          float64 `mapstructure:"request_cost"`             // Cost per request in USD
	Support              string  `mapstructure:"support"`                  // Support level description
	HistoricalData       bool    `mapstructure:"historical_data"`          // Access to historical data
	PrivateData          bool    `mapstructure:"private_data"`             // Access to private data feeds
}

type Config struct {
	PricePoolTTL         int                         `mapstructure:"price_pool_ttl"`
	RELAY_TIME_THRESHOLD int                         `mapstructure:"RELAY_TIME_THRESHOLD"`
	AggregatorNodes      int                         `mapstructure:"aggregator_nodes"`
	ConsensusThresh      float64                     `mapstructure:"consensus_threshold"`
	AggrDevPerc          float32                     `mapstructure:"aggr_dev_perc"`
	Assets               []AssetConfig               `mapstructure:"assets"`
	ApiKeys              ApiKey                      `mapstructure:"api_keys"`
	Contracts            []ContractConfig            `mapstructure:"contracts"`
	PrivateKey           string                      `mapstructure:"private_key"`
	DB_URL               string                      `mapstructure:"DB_URL"`
	SERVER_PORT          string                      `mapstructure:"server_port"`
	JWTSecret            string                      `mapstructure:"jwt_secret"`
	SubscriptionPlans    map[string]SubscriptionPlan `mapstructure:"subscription_plans"`
	IFALabsAPIURL        string                      `mapstructure:"ifa_labs_api_url"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetDefault("price_pool_ttl", 10)
	viper.SetDefault("aggregator_nodes", 3)
	viper.SetDefault("consensus_threshold", 0.01)
	viper.SetDefault("ifa_labs_api_url", os.Getenv("IFA_LABS_API_URL"))
	viper.SetDefault("api_keys", map[string]string{
		"monierate": os.Getenv("MONIERATE_API_KEY"),
		"ifalabs":   os.Getenv("IFA_LABS_API_KEY"),
	})
	viper.SetDefault("subscription_plans", map[string]SubscriptionPlan{
		"free": {
			Name:             "Free tier",
			Price:            0,
			APIRequests:      1000,
			RateLimitPerHour: 10,  // 10 requests per hour
			RateLimitPerDay:  100, // 100 requests per day
			DataAccess:       "Two feeds",
			CustomPairs:      0,
			RequestCost:      0,
			Support:          "Email & Community",
			HistoricalData:   false,
			PrivateData:      false,
		},
		"developer": {
			Name:             "Developer tier",
			Price:            50,
			APIRequests:      10000,
			RateLimitPerHour: 100,  // 100 requests per hour
			RateLimitPerDay:  1000, // 1000 requests per day
			DataAccess:       "All feeds",
			CustomPairs:      0,
			RequestCost:      0.0005,
			Support:          "24/7 support",
			HistoricalData:   false,
			PrivateData:      false,
		},
		"professional": {
			Name:             "Professional tier",
			Price:            100,
			APIRequests:      100000,
			RateLimitPerHour: 500,   // 500 requests per hour
			RateLimitPerDay:  10000, // 10,000 requests per day
			DataAccess:       "All feeds + Historical data",
			CustomPairs:      3,
			RequestCost:      0.0002,
			Support:          "24/7 support",
			HistoricalData:   true,
			PrivateData:      false,
		},
		"enterprise": {
			Name:             "Enterprise tier",
			Price:            0, // Custom pricing
			APIRequests:      0, // Unlimited
			RateLimitPerHour: 0, // Unlimited
			RateLimitPerDay:  0, // Unlimited
			DataAccess:       "All feeds + Private",
			CustomPairs:      -1, // Custom/unlimited
			RequestCost:      0,  // Custom
			Support:          "24/7 support + dedicated engineer",
			HistoricalData:   true,
			PrivateData:      true,
		},
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

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = "your-secret-key-here" // Default secret (not secure for production)
		}
	}


	return &cfg
}
