package timescale

import (
	"context"
	"encoding/json"
	"fmt"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type TimescaleGORM struct {
	db *gorm.DB
}

// NewTimescaleGORM creates a new TimescaleDB instance with GORM
func NewTimescaleGORM(connStr string) (*TimescaleGORM, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	ts := &TimescaleGORM{db: db}
	if err := ts.Initialize(context.Background()); err != nil {
		return nil, err
	}
	return ts, nil
}

// Initialize sets up the database schema and hypertables
func (t *TimescaleGORM) Initialize(ctx context.Context) error {
	// Create extensions
	if err := t.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Handle migration for removing key_encrypted column
	if err := t.handleKeyEncryptedMigration(); err != nil {
		logging.Logger.Warn("Failed to handle key_encrypted migration", zap.Error(err))
	}

	// Auto-migrate tables
	if err := t.db.AutoMigrate(&Price{}, &RawPrice{}, &PriceRawPriceLink{}, &Issuance{},
		&CompanyProfile{}, &DashboardAPIKey{}, &DashboardAPIKeyUsage{}, &DashboardPayment{}, &VerificationToken{}, &Invoice{}); err != nil {
		return err
	}

	// Create hypertable for prices (TimescaleDB specific)
	if err := t.db.Exec("SELECT create_hypertable('prices', 'timestamp', if_not_exists => true, create_default_indexes => false)").Error; err != nil {
		logging.Logger.Warn("Failed to create hypertable, might already exist", zap.Error(err))
	}

	// Create additional indexes
	if err := t.db.Exec("CREATE INDEX IF NOT EXISTS idx_prices_id ON prices(id)").Error; err != nil {
		logging.Logger.Warn("Failed to create index", zap.Error(err))
	}

	// Create invoice table indexes and constraints
	if err := t.setupInvoiceTable(); err != nil {
		logging.Logger.Warn("Failed to setup invoice table", zap.Error(err))
	}

	logging.Logger.Info("Database tables initialized with GORM")
	return nil
}

// SavePrice saves a unified price to the database
func (t *TimescaleGORM) SavePrice(ctx context.Context, price models.UnifiedPrice) error {
	gormPrice := Price{
		ID:        uuid.MustParse(price.ID),
		AssetID:   price.AssetID,
		Value:     price.Value,
		Expo:      price.Expo,
		Timestamp: price.Timestamp,
		Source:    price.Source,
		ReqHash:   price.ReqHash,
	}

	return t.db.WithContext(ctx).Create(&gormPrice).Error
}

// GetLastPrice retrieves the most recent price for an asset
func (t *TimescaleGORM) GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error) {
	var price Price
	err := t.db.WithContext(ctx).
		Where("asset_id = ?", assetID).
		Order("timestamp DESC").
		First(&price).Error

	if err != nil {
		return nil, err
	}

	return &models.UnifiedPrice{
		ID:        price.ID.String(),
		AssetID:   price.AssetID,
		Value:     price.Value,
		Expo:      price.Expo,
		Timestamp: price.Timestamp,
		Source:    price.Source,
		ReqHash:   price.ReqHash,
	}, nil
}

// SaveRawPrice saves raw price data from feeds
func (t *TimescaleGORM) SaveRawPrice(ctx context.Context, price models.Price) error {
	rawPrice := RawPrice{
		ID:        price.ID,
		Source:    price.Source,
		ReqURL:    price.ReqURL,
		AssetID:   price.InternalAssetIdentity,
		Value:     price.Value,
		Expo:      price.Expo,
		Timestamp: price.Timestamp,
	}

	return t.db.WithContext(ctx).Create(&rawPrice).Error
}

// LinkRawPricesToAggregatedPrice links raw prices to an aggregated price
func (t *TimescaleGORM) LinkRawPricesToAggregatedPrice(ctx context.Context, aggregatedPriceID string, timestamp time.Time, rawPriceIDs []string) error {
	// Verify the aggregated price exists
	var count int64
	err := t.db.WithContext(ctx).Model(&Price{}).
		Where("id = ? AND timestamp = ?", aggregatedPriceID, timestamp).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("aggregated price not found: %s at %s", aggregatedPriceID, timestamp)
	}

	// Filter out empty raw price IDs
	filtered := make([]string, 0, len(rawPriceIDs))
	for _, rawID := range rawPriceIDs {
		if strings.TrimSpace(rawID) != "" {
			filtered = append(filtered, rawID)
		}
	}

	logging.Logger.Info("Linking filtered raw prices", zap.Any("filtered_link", filtered))

	// Create links
	for _, rawID := range filtered {
		link := PriceRawPriceLink{
			PriceID:        uuid.MustParse(aggregatedPriceID),
			PriceTimestamp: timestamp,
			RawPriceID:     rawID,
		}

		err := t.db.WithContext(ctx).
			Where("price_id = ? AND price_timestamp = ? AND raw_price_id = ?",
				link.PriceID, link.PriceTimestamp, link.RawPriceID).
			FirstOrCreate(&link).Error
		if err != nil {
			logging.Logger.Error("Failed to create price link",
				zap.Error(err),
				zap.String("raw_price_id", rawID),
				zap.String("aggregated_price_id", aggregatedPriceID),
				zap.Time("timestamp", timestamp))
			return err
		}
	}

	return nil
}

// SaveIssuance saves an issuance to the database
func (t *TimescaleGORM) SaveIssuance(ctx context.Context, issuance models.Issuance) error {
	// Save price if approved
	if issuance.State == models.Approved {
		if err := t.SavePrice(ctx, issuance.Price); err != nil {
			logging.Logger.Error("Error saving price", zap.Error(err), zap.String("price_id", issuance.Price.ID))
			return err
		}
	}

	jsonMetadata, err := json.Marshal(issuance.Metadata)
	if err != nil {
		return err
	}
	gormIssuance := Issuance{
		ID:             issuance.ID,
		State:          int16(issuance.State),
		IssuerAddress:  issuance.IssuerAddress,
		RoundID:        int64(issuance.RoundID),
		CreatedAt:      issuance.CreatedAt,
		UpdatedAt:      issuance.UpdatedAt,
		PriceValue:     issuance.PriceValue,
		PriceAssetID:   issuance.PriceAssetID,
		PriceSource:    issuance.PriceSource,
		PriceTimestamp: issuance.PriceTimestamp,
		Metadata:       datatypes.JSON(jsonMetadata),
	}
	return t.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"state":      gorm.Expr("EXCLUDED.state"),
				"updated_at": gorm.Expr("EXCLUDED.updated_at"),
				"metadata":   gorm.Expr("EXCLUDED.metadata"),
			}),
		}).
		Create(&gormIssuance).Error
}

// GetLastIssuance retrieves the most recent issuance for an asset
func (t *TimescaleGORM) GetLastIssuance(ctx context.Context, assetID string) (*models.Issuance, error) {
	var issuance Issuance
	err := t.db.WithContext(ctx).
		Where("price_asset_id = ?", assetID).
		Order("created_at DESC").
		First(&issuance).Error

	if err != nil {
		return nil, err
	}

	return &models.Issuance{
		ID:             issuance.ID,
		State:          models.IssuanceState(issuance.State),
		IssuerAddress:  issuance.IssuerAddress,
		RoundID:        uint64(issuance.RoundID),
		CreatedAt:      issuance.CreatedAt,
		UpdatedAt:      issuance.UpdatedAt,
		PriceValue:     issuance.PriceValue,
		PriceAssetID:   issuance.PriceAssetID,
		PriceSource:    issuance.PriceSource,
		PriceTimestamp: issuance.PriceTimestamp,
		Metadata:       issuance.Metadata,
	}, nil
}

// GetIssuance retrieves an issuance by ID
func (t *TimescaleGORM) GetIssuance(ctx context.Context, id string) (*models.Issuance, error) {
	var issuance Issuance
	err := t.db.WithContext(ctx).
		Where("id = ?", id).
		First(&issuance).Error

	if err != nil {
		return nil, err
	}

	return &models.Issuance{
		ID:             issuance.ID,
		State:          models.IssuanceState(issuance.State),
		IssuerAddress:  issuance.IssuerAddress,
		RoundID:        uint64(issuance.RoundID),
		CreatedAt:      issuance.CreatedAt,
		UpdatedAt:      issuance.UpdatedAt,
		PriceValue:     issuance.PriceValue,
		PriceAssetID:   issuance.PriceAssetID,
		PriceSource:    issuance.PriceSource,
		PriceTimestamp: issuance.PriceTimestamp,
		Metadata:       issuance.Metadata,
	}, nil
}

// AuditPrice retrieves a price and all its associated raw prices
func (t *TimescaleGORM) AuditPrice(ctx context.Context, id string) (*models.PriceAudit, error) {
	// Get the aggregated price
	var price Price
	err := t.db.WithContext(ctx).
		Where("id = ?", id).
		Order("timestamp DESC").
		First(&price).Error
	if err != nil {
		return nil, err
	}

	// Get associated raw prices
	var rawPrices []RawPrice
	err = t.db.WithContext(ctx).
		Joins("JOIN price_raw_price_links l ON raw_prices.id = l.raw_price_id").
		Where("l.price_id = ?", id).
		Order("raw_prices.timestamp").
		Find(&rawPrices).Error
	if err != nil {
		return nil, err
	}

	// Convert to models
	unifiedPrice := models.UnifiedPrice{
		ID:        price.ID.String(),
		AssetID:   price.AssetID,
		Value:     price.Value,
		Expo:      price.Expo,
		Timestamp: price.Timestamp,
		Source:    price.Source,
		ReqHash:   price.ReqHash,
	}

	var modelRawPrices []models.Price
	for _, rp := range rawPrices {
		modelRawPrices = append(modelRawPrices, models.Price{
			ID:                    rp.ID,
			Source:                rp.Source,
			ReqURL:                rp.ReqURL,
			InternalAssetIdentity: rp.AssetID,
			Value:                 rp.Value,
			Expo:                  rp.Expo,
			Timestamp:             rp.Timestamp,
		})
	}

	return &models.PriceAudit{
		PriceID:         price.ID.String(),
		AssetID:         price.AssetID,
		AggregatedPrice: unifiedPrice,
		RawPrices:       modelRawPrices,
		CreatedAt:       price.Timestamp,
		UpdatedAt:       price.Timestamp,
	}, nil
}

// GetHistoricalPrice retrieves historical price data
func (t *TimescaleGORM) GetHistoricalPrice(ctx context.Context, assetID string, lookback time.Duration) (*models.UnifiedPrice, error) {
	historicalTime := time.Now().Add(-lookback)

	var price Price
	err := t.db.WithContext(ctx).
		Where("asset_id = ? AND timestamp <= ?", assetID, historicalTime).
		Order("timestamp DESC").
		First(&price).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No historical price found
		}
		return nil, err
	}

	return &models.UnifiedPrice{
		ID:        price.ID.String(),
		AssetID:   price.AssetID,
		Value:     price.Value,
		Expo:      price.Expo,
		Timestamp: price.Timestamp,
		Source:    price.Source,
		ReqHash:   price.ReqHash,
	}, nil
}

// GetRawPricesForAsset retrieves all raw prices for a specific asset within a time range
func (t *TimescaleGORM) GetRawPricesForAsset(ctx context.Context, assetID string, start, end time.Time) ([]models.Price, error) {
	var rawPrices []RawPrice
	err := t.db.WithContext(ctx).
		Where("asset_id = ? AND timestamp BETWEEN ? AND ?", assetID, start, end).
		Order("timestamp").
		Find(&rawPrices).Error

	if err != nil {
		return nil, err
	}

	var result []models.Price
	for _, rp := range rawPrices {
		result = append(result, models.Price{
			ID:                    rp.ID,
			Source:                rp.Source,
			ReqURL:                rp.ReqURL,
			InternalAssetIdentity: rp.AssetID,
			Value:                 rp.Value,
			Expo:                  rp.Expo,
			Timestamp:             rp.Timestamp,
		})
	}

	return result, nil
}

// GetPriceHistory retrieves price history for an asset
func (t *TimescaleGORM) GetPriceHistory(ctx context.Context, assetID string, start, end time.Time, limit int) ([]models.UnifiedPrice, error) {
	var prices []Price
	query := t.db.WithContext(ctx).
		Where("asset_id = ? AND timestamp BETWEEN ? AND ?", assetID, start, end).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&prices).Error
	if err != nil {
		return nil, err
	}

	var result []models.UnifiedPrice
	for _, p := range prices {
		result = append(result, models.UnifiedPrice{
			ID:        p.ID.String(),
			AssetID:   p.AssetID,
			Value:     p.Value,
			Expo:      p.Expo,
			Timestamp: p.Timestamp,
			Source:    p.Source,
			ReqHash:   p.ReqHash,
		})
	}

	return result, nil
}

// create account
// func (t *TimescaleGORM) CreateDashboardAccount(ctx context.Context, profile G) {}

// Close closes the database connection
func (t *TimescaleGORM) Close() error {
	sqlDB, err := t.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the GORM database instance for dashboard operations
func (t *TimescaleGORM) GetDB() *gorm.DB {
	return t.db
}

// handleKeyEncryptedMigration removes the key_encrypted column if it exists and ensures key_plain exists
func (t *TimescaleGORM) handleKeyEncryptedMigration() error {
	// Check if the key_encrypted column exists
	var encryptedColumnExists bool
	err := t.db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dashboard_api_keys' 
			AND column_name = 'key_encrypted'
		)
	`).Scan(&encryptedColumnExists).Error
	
	if err != nil {
		return fmt.Errorf("failed to check for key_encrypted column: %w", err)
	}
	
	// If the key_encrypted column exists, drop it
	if encryptedColumnExists {
		logging.Logger.Info("Dropping key_encrypted column from dashboard_api_keys table")
		if err := t.db.Exec("ALTER TABLE dashboard_api_keys DROP COLUMN IF EXISTS key_encrypted").Error; err != nil {
			return fmt.Errorf("failed to drop key_encrypted column: %w", err)
		}
		logging.Logger.Info("Successfully dropped key_encrypted column")
	}
	
	// Check if the key_plain column exists
	var plainColumnExists bool
	err = t.db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'dashboard_api_keys' 
			AND column_name = 'key_plain'
		)
	`).Scan(&plainColumnExists).Error
	
	if err != nil {
		return fmt.Errorf("failed to check for key_plain column: %w", err)
	}
	
	// If the key_plain column doesn't exist, add it
	if !plainColumnExists {
		logging.Logger.Info("Adding key_plain column to dashboard_api_keys table")
		if err := t.db.Exec("ALTER TABLE dashboard_api_keys ADD COLUMN key_plain text NOT NULL DEFAULT ''").Error; err != nil {
			return fmt.Errorf("failed to add key_plain column: %w", err)
		}
		logging.Logger.Info("Successfully added key_plain column")
		
		// Copy existing key_hash values to key_plain for existing records
		logging.Logger.Info("Copying existing key_hash values to key_plain")
		if err := t.db.Exec("UPDATE dashboard_api_keys SET key_plain = key_hash WHERE key_plain = ''").Error; err != nil {
			logging.Logger.Warn("Failed to copy key_hash to key_plain", zap.Error(err))
		}
	}
	
	return nil
}

// setupInvoiceTable creates indexes and constraints for the invoice table
func (t *TimescaleGORM) setupInvoiceTable() error {
	// Add foreign key constraint to company_profiles table
	if err := t.db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE constraint_name = 'fk_invoices_account_id' 
				AND table_name = 'invoices'
			) THEN
				ALTER TABLE invoices 
				ADD CONSTRAINT fk_invoices_account_id 
				FOREIGN KEY (account_id) REFERENCES company_profiles(id) ON DELETE CASCADE;
			END IF;
		END $$;
	`).Error; err != nil {
		logging.Logger.Warn("Failed to add foreign key constraint for invoices", zap.Error(err))
	}

	// Add unique index to prevent duplicate invoices per account per due date
	if err := t.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_account_due_date 
		ON invoices (account_id, due_date) 
		WHERE status != 'void';
	`).Error; err != nil {
		logging.Logger.Warn("Failed to create unique index for invoices", zap.Error(err))
	}

	// Add performance indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_invoices_account_id ON invoices (account_id);",
		"CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices (status);",
		"CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices (due_date);",
		"CREATE INDEX IF NOT EXISTS idx_invoices_issued_at ON invoices (issued_at);",
	}

	for _, indexSQL := range indexes {
		if err := t.db.Exec(indexSQL).Error; err != nil {
			logging.Logger.Warn("Failed to create invoice index", zap.Error(err), zap.String("sql", indexSQL))
		}
	}

	// Add trigger to automatically update updated_at timestamp
	if err := t.db.Exec(`
		CREATE OR REPLACE FUNCTION update_invoices_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`).Error; err != nil {
		logging.Logger.Warn("Failed to create update_invoices_updated_at function", zap.Error(err))
	}

	if err := t.db.Exec(`
		DROP TRIGGER IF EXISTS trigger_update_invoices_updated_at ON invoices;
		CREATE TRIGGER trigger_update_invoices_updated_at
			BEFORE UPDATE ON invoices
			FOR EACH ROW
			EXECUTE FUNCTION update_invoices_updated_at();
	`).Error; err != nil {
		logging.Logger.Warn("Failed to create invoice update trigger", zap.Error(err))
	}

	logging.Logger.Info("Invoice table setup completed")
	return nil
}
