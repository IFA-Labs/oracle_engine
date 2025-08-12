package timescale

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Price represents the aggregated/unified price table
type Price struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AssetID   string    `gorm:"type:text;not null;index" json:"asset_id"`
	Value     float64   `gorm:"type:float8;not null" json:"value"`
	Expo      int8      `gorm:"type:smallint;not null" json:"expo"`
	Timestamp time.Time `gorm:"type:timestamptz;not null;primaryKey" json:"timestamp"`
	Source    string    `gorm:"type:text;not null" json:"source"`
	ReqHash   string    `gorm:"type:text" json:"req_hash"`

	// Relationships
	RawPriceLinks []PriceRawPriceLink `gorm:"foreignKey:PriceID,PriceTimestamp;references:ID,Timestamp" json:"raw_price_links,omitempty"`
}

// TableName specifies the table name for the Price model
func (Price) TableName() string {
	return "prices"
}

// BeforeCreate sets the ID if not already set
func (p *Price) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// RawPrice represents the raw price data from individual feeds
type RawPrice struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	Source    string    `gorm:"type:text;not null" json:"source"`
	ReqURL    string    `gorm:"type:text" json:"req_url"`
	AssetID   string    `gorm:"type:text;not null" json:"asset_id"`
	Value     float64   `gorm:"type:float8;not null" json:"value"`
	Expo      int8      `gorm:"type:smallint;not null" json:"expo"`
	Timestamp time.Time `gorm:"type:timestamptz;not null" json:"timestamp"`

	// Relationships
	PriceLinks []PriceRawPriceLink `gorm:"foreignKey:RawPriceID" json:"price_links,omitempty"`
}

// TableName specifies the table name for the RawPrice model
func (RawPrice) TableName() string {
	return "raw_prices"
}

// PriceRawPriceLink represents the linking table between prices and raw prices
type PriceRawPriceLink struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PriceID        uuid.UUID `gorm:"type:uuid;not null" json:"price_id"`
	PriceTimestamp time.Time `gorm:"type:timestamptz;not null" json:"price_timestamp"`
	RawPriceID     string    `gorm:"type:text;not null" json:"raw_price_id"`

	// Relationships
	Price    Price    `gorm:"foreignKey:PriceID,PriceTimestamp;references:ID,Timestamp" json:"price,omitempty"`
	RawPrice RawPrice `gorm:"foreignKey:RawPriceID" json:"raw_price,omitempty"`
}

// TableName specifies the table name for the PriceRawPriceLink model
func (PriceRawPriceLink) TableName() string {
	return "price_raw_price_links"
}

// Issuance represents the issuance table
type Issuance struct {
	ID             string         `gorm:"type:text;primaryKey" json:"id"`
	State          int16          `gorm:"type:smallint;not null" json:"state"`
	IssuerAddress  string         `gorm:"type:text;not null" json:"issuer_address"`
	RoundID        int64          `gorm:"type:bigint;not null" json:"round_id"`
	CreatedAt      time.Time      `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	PriceValue     float64        `gorm:"type:float8;not null" json:"price_value"`
	PriceAssetID   string         `gorm:"type:text;not null" json:"price_asset_id"`
	PriceSource    string         `gorm:"type:text;not null" json:"price_source"`
	PriceTimestamp time.Time      `gorm:"type:timestamptz;not null" json:"price_timestamp"`
	Metadata       datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
}

// TableName specifies the table name for the Issuance model
func (Issuance) TableName() string {
	return "issuances"
}

type CompanyProfile struct {
	ID          string    `gorm:"type:text;primaryKey" json:"id"`
	Name        string    `gorm:"type:text;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	LogoURL     string    `gorm:"type:text" json:"logo_url"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null" json:"updated_at"`
	// merging user and company profile for now
	FirstName string `gorm:"type:text;not null" json:"first_name"`
	LastName  string `gorm:"type:text;not null" json:"last_name"`
	Password  string `gorm:"type:text;not null" json:"password"`
	Email     string `gorm:"type:text;not null" json:"email"`
}

func (CompanyProfile) TableName() string {
	return "company_profiles"
}

type DashboardAPIKey struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	ProfileID string    `gorm:"type:text;not null" json:"profile_id"`
	Key       string    `gorm:"type:text;not null" json:"key"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null" json:"updated_at"`
}

func (DashboardAPIKey) TableName() string {
	return "dashboard_api_keys"
}

// track user's usage with api keys
type DashboardAPIKeyUsage struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	ProfileID string    `gorm:"type:text;not null" json:"profile_id"`
	KeyID     string    `gorm:"type:text;not null" json:"key_id"`
	Endpoint  string    `gorm:"type:text;not null" json:"endpoint"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null" json:"updated_at"`
}

func (DashboardAPIKeyUsage) TableName() string {
	return "dashboard_api_key_usages"
}

// handle billing and payment
type DashboardPayment struct {
	ID               string    `gorm:"type:text;primaryKey" json:"id"`
	ProfileID        string    `gorm:"type:text;not null" json:"profile_id"`
	Amount           float64   `gorm:"type:float8;not null" json:"amount"`
	Status           string    `gorm:"type:text;not null" json:"status"` //
	CreatedAt        time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;not null" json:"updated_at"`
	SubscriptionType string    `gorm:"type:text;not null" json:"subscription_type"` // e.g., "basic", "premium"
}
