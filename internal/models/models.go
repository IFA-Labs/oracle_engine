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

type PriceChange struct {
	Period    string    `json:"period"`     // e.g. "7d", "3d", "24h"
	Change    float64   `json:"change"`     // Absolute change
	ChangePct float64   `json:"change_pct"` // Percentage change
	FromPrice float64   `json:"from_price"` // Starting price
	ToPrice   float64   `json:"to_price"`   // Current price
	FromTime  time.Time `json:"from_time"`  // Starting time
	ToTime    time.Time `json:"to_time"`    // Current time
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
	IsAggr            bool          `json:"is_aggr"`
	ConnectedPriceIDs []string      `json:"connected_price_ids"`
	PriceChanges      []PriceChange `json:"price_changes,omitempty"` // Optional price changes
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
	Asset   string `json:"asset"`
}

// CalculatePriceChange calculates the price change between two prices
func CalculatePriceChange(current, historical *UnifiedPrice, period string) *PriceChange {
	if historical == nil {
		return nil
	}

	currentNum := current.Number()
	historicalNum := historical.Number()

	change := currentNum - historicalNum
	changePct := (change / historicalNum) * 100

	return &PriceChange{
		Period:    period,
		Change:    change,
		ChangePct: changePct,
		FromPrice: historicalNum,
		ToPrice:   currentNum,
		FromTime:  historical.Timestamp,
		ToTime:    current.Timestamp,
	}
}

// Dashboard domain models
type CompanyProfile struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Website          string    `json:"website"`
	LogoURL          string    `json:"logo_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	Password         string    `json:"-"`                 // dont return password to user
	SubscriptionPlan string    `json:"subscription_plan"` // "free", "developer", "professional", "enterprise"
}

type SignUpRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Website     string `json:"website" binding:"omitempty,url"`
	Description string `json:"description"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	// company name
	Name string `json:"name" binding:"required"`
}

type SignUpResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type UpdateProfileRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Website     *string `json:"website,omitempty" binding:"omitempty,url"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty" binding:"omitempty,url"`
}

type APIKey struct {
	ID               string     `json:"id"`
	Key              string     `json:"key,omitempty"` // Only show on creation
	KeyHash          string     `json:"-"`             // Never expose
	ProfileID        string     `json:"profile_id"`
	SubscriptionPlan string     `json:"subscription_plan,omitempty"` // Added for rate limiting context
	Name             string     `json:"name"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastUsed         *time.Time `json:"last_used,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type CreateAPIKeyResponse struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

type APIKeyUsage struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	ProfileID string    `json:"profile_id"`
	Endpoint  string    `json:"endpoint"`
	Method    string    `json:"method"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID               string    `json:"id"`
	ProfileID        string    `json:"profile_id"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	Status           string    `json:"status"` // "pending", "completed", "failed", "refunded"
	SubscriptionType string    `json:"subscription_type"`
	PaymentMethod    string    `json:"payment_method"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreatePaymentRequest struct {
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	Currency         string  `json:"currency" binding:"required,oneof=USD EUR GBP"`
	SubscriptionType string  `json:"subscription_type" binding:"required,oneof=basic premium enterprise"`
	PaymentMethod    string  `json:"payment_method" binding:"required,oneof=card stripe paypal"`
}

type PaymentHistoryResponse struct {
	Payments   []Payment `json:"payments"`
	TotalCount int64     `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

// Subscription plan limits
type SubscriptionPlan struct {
	Name           string  `json:"name"`
	Price          float64 `json:"price"`           // Monthly price in USD
	APIRequests    int64   `json:"api_requests"`    // Monthly API request limit (0 = unlimited)
	RateLimit      int     `json:"rate_limit"`      // Rate limit in hours
	DataAccess     string  `json:"data_access"`     // Description of data access level
	CustomPairs    int     `json:"custom_pairs"`    // Number of custom pairs allowed
	RequestCost    float64 `json:"request_cost"`    // Cost per request in USD
	Support        string  `json:"support"`         // Support level description
	HistoricalData bool    `json:"historical_data"` // Access to historical data
	PrivateData    bool    `json:"private_data"`    // Access to private data feeds
}

// API Usage tracking with subscription context
type APIKeyUsageStats struct {
	KeyID            string    `json:"key_id"`
	ProfileID        string    `json:"profile_id"`
	SubscriptionPlan string    `json:"subscription_plan"`
	MonthlyUsage     int64     `json:"monthly_usage"`     // Current month usage
	MonthlyLimit     int64     `json:"monthly_limit"`     // Monthly limit based on plan
	DailyUsage       int64     `json:"daily_usage"`       // Today's usage
	LastRequestTime  time.Time `json:"last_request_time"` // For rate limiting
	IsLimitExceeded  bool      `json:"is_limit_exceeded"`
}
