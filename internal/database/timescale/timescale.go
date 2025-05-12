package timescale

import (
	"context"
	"database/sql"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type TimescaleDB struct {
	db *sql.DB
}

func NewTimescaleDB(connStr string) (*TimescaleDB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	ts := &TimescaleDB{db: db}
	if err := ts.Initialize(context.Background()); err != nil {
		return nil, err
	}
	return ts, nil
}

func (t *TimescaleDB) Initialize(ctx context.Context) error {
	query := `
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS prices (
		id UUID DEFAULT uuid_generate_v4(),
		asset_id TEXT NOT NULL,
		value FLOAT8 NOT NULL,
		expo SMALLINT NOT NULL,
		timestamp TIMESTAMPTZ NOT NULL,                  
		source TEXT NOT NULL,
		req_hash TEXT,
		PRIMARY KEY (id, timestamp)
	);

	SELECT create_hypertable('prices', 'timestamp', if_not_exists => true, create_default_indexes => false);
	CREATE INDEX ON prices(id);

    CREATE TABLE IF NOT EXISTS raw_prices (
        id TEXT PRIMARY KEY,
        source TEXT NOT NULL,
        req_url TEXT,
		asset_id TEXT NOT NULL,
        value FLOAT8 NOT NULL,
        expo SMALLINT NOT NULL,
        timestamp TIMESTAMPTZ NOT NULL
    );

    CREATE TABLE IF NOT EXISTS price_raw_price_links (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		price_id UUID NOT NULL,
		price_timestamp TIMESTAMPTZ NOT NULL,
		raw_price_id TEXT NOT NULL,
		-- PRIMARY KEY (price_id, raw_price_id)

		FOREIGN KEY (price_id, price_timestamp) REFERENCES prices(id, timestamp) ON DELETE CASCADE,
		FOREIGN KEY (raw_price_id) REFERENCES raw_prices(id) ON DELETE CASCADE
	);


    CREATE TABLE IF NOT EXISTS issuances (
        id TEXT PRIMARY KEY,
        state SMALLINT NOT NULL,
        issuer_address TEXT NOT NULL,
        round_id BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL,
        price_value FLOAT8 NOT NULL,
        price_asset_id TEXT NOT NULL,
        price_source TEXT NOT NULL,
        price_timestamp TIMESTAMPTZ NOT NULL,
        metadata JSONB
    );
	`
	_, err := t.db.ExecContext(ctx, query)
	if err != nil {
		logging.Logger.Error("Failed to initialize database tables", zap.Error(err))
		return err
	}
	logging.Logger.Info("Database tables initialized")
	return nil
}

func (t *TimescaleDB) SavePrice(ctx context.Context, price models.UnifiedPrice) error {
	query := `
        INSERT INTO prices (id, asset_id, value, expo, timestamp, source, req_hash)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := t.db.ExecContext(ctx, query,
		price.ID, price.AssetID, price.Value, price.Expo, price.Timestamp, price.Source, price.ReqHash)
	return err
}

func (t *TimescaleDB) GetLastPrice(ctx context.Context, assetID string) (*models.UnifiedPrice, error) {
	query := `
        SELECT value, expo, timestamp, source, req_hash
        FROM prices
        WHERE asset_id = $1
        ORDER BY timestamp DESC
        LIMIT 1`
	var value float64
	var expo int8
	var timestamp time.Time
	var source, req_hash string
	err := t.db.QueryRowContext(ctx, query, assetID).Scan(&value, &expo, &timestamp, &source, &req_hash)
	if err != nil {
		return nil, err
	}
	logging.Logger.Warn("caught", zap.Float64("key", value), zap.Int32("expo", int32(expo)))
	return &models.UnifiedPrice{
		Value:     value,
		Expo:      expo,
		Timestamp: timestamp,
		ReqHash:   req_hash,
		Source:    source,
	}, nil
}

func (t *TimescaleDB) SaveIssuance(ctx context.Context, issuance models.Issuance) error {
	if issuance.State == models.Approved {
		if err := t.SavePrice(ctx, issuance.Price); err != nil {
			logging.Logger.Info("Error saving price", zap.Any("err", err), zap.Any("price", issuance.Price.ID))
			return err
		}
	}

	query := `
        INSERT INTO issuances (
            id, state, issuer_address, round_id, created_at, updated_at,
            price_value, price_asset_id, price_source, price_timestamp,
            metadata
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (id) DO UPDATE SET
            state = EXCLUDED.state,
            updated_at = EXCLUDED.updated_at,
            metadata = EXCLUDED.metadata
    `
	_, err := t.db.ExecContext(ctx, query,
		issuance.ID,
		issuance.State,
		issuance.IssuerAddress,
		issuance.RoundID,
		issuance.CreatedAt,
		issuance.UpdatedAt,
		issuance.PriceValue,
		issuance.PriceAssetID,
		issuance.PriceSource,
		issuance.PriceTimestamp,
		issuance.Metadata,
	)
	return err
}

func (t *TimescaleDB) GetIssuance(ctx context.Context, id string) (*models.Issuance, error) {
	query := `
        SELECT 
            id, state, issuer_address, round_id, created_at, updated_at,
            price_value, price_asset_id, price_source, price_timestamp,
            metadata
        FROM issuances
        WHERE id = $1
    `
	var issuance models.Issuance
	err := t.db.QueryRowContext(ctx, query, id).Scan(
		&issuance.ID,
		&issuance.State,
		&issuance.IssuerAddress,
		&issuance.RoundID,
		&issuance.CreatedAt,
		&issuance.UpdatedAt,
		&issuance.PriceValue,
		&issuance.PriceAssetID,
		&issuance.PriceSource,
		&issuance.PriceTimestamp,
		&issuance.Metadata,
	)
	if err != nil {
		return nil, err
	}
	return &issuance, nil
}

func (t *TimescaleDB) SaveRawPrice(ctx context.Context, price models.Price) error {
	query := `
        INSERT INTO raw_prices (id, source, req_url, asset_id, value, expo, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := t.db.ExecContext(ctx, query,
		price.ID,
		price.Source,
		price.ReqURL,
		price.InternalAssetIdentity,
		price.Value,
		price.Expo,
		price.Timestamp,
	)
	return err
}

func (t *TimescaleDB) LinkRawPricesToAggregatedPrice(ctx context.Context, aggregatedPriceID string, timestamp time.Time, rawPriceIDs []string) error {

	existsQuery := `SELECT 1 FROM prices WHERE id = $1 AND timestamp = $2`
	var dummy int
	err := t.db.QueryRowContext(ctx, existsQuery, aggregatedPriceID, timestamp).Scan(&dummy)
	if err == sql.ErrNoRows {
		logging.Logger.Info("Cutt", zap.String("ai", aggregatedPriceID), zap.String("ti", timestamp.String()))
	} else if err != nil {
		return err
	}

	query := `
        INSERT INTO price_raw_price_links (price_id, price_timestamp, raw_price_id)
        VALUES ($1, $2, $3)
        ON CONFLICT DO NOTHING
    `
	filtered := make([]string, 0, len(rawPriceIDs))
	for _, rawID := range rawPriceIDs {
		if strings.TrimSpace(rawID) != "" {
			filtered = append(filtered, rawID)
		}
	}

	logging.Logger.Info("Linking filtered raw prices", zap.Any("filtered_link", filtered))

	for _, rawID := range filtered {
		_, err := t.db.ExecContext(ctx, query, aggregatedPriceID, timestamp, rawID)
		if err != nil {
			logging.Logger.Info(
				"Killllllll",
				zap.Any("errrrr", err), zap.Any("str", rawID),
				zap.String("agg", aggregatedPriceID),
				zap.Time("time", timestamp),
			)
			return err
		}
	}
	return nil
}

func (t *TimescaleDB) AuditPrice(ctx context.Context, id string) (*models.PriceAudit, error) {
	priceQuery := `
        SELECT id, asset_id, value, expo, timestamp, source, req_hash
        FROM prices
        WHERE id = $1
        ORDER BY timestamp DESC
        LIMIT 1
    `
	var up models.UnifiedPrice
	err := t.db.QueryRowContext(ctx, priceQuery, id).Scan(
		&up.ID, &up.AssetID, &up.Value, &up.Expo, &up.Timestamp, &up.Source, &up.ReqHash,
	)
	if err != nil {
		return nil, err
	}

	// rawQuery := `
	// 	SELECT r.id, r.source, r.req_url, r.asset_id, r.value, r.expo, r.timestamp
	// 	FROM raw_prices r
	// 	INNER JOIN price_raw_price_links l
	// 		ON r.id = l.raw_price_id
	// 	WHERE l.price_id = $1
	// `

	rawQuery := `
		SELECT r.id, r.source, r.req_url, r.asset_id, r.value, r.expo, r.timestamp
		FROM price_raw_price_links l
		INNER JOIN raw_prices r ON r.id = l.raw_price_id
		WHERE l.price_id = $1
		ORDER BY r.timestamp;
	`

	rows, err := t.db.QueryContext(ctx, rawQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var raws []models.Price
	for rows.Next() {
		var rp models.Price
		err := rows.Scan(&rp.ID, &rp.Source, &rp.ReqURL, &rp.InternalAssetIdentity, &rp.Value, &rp.Expo, &rp.Timestamp)
		if err != nil {
			return nil, err
		}
		raws = append(raws, rp)
	}

	auditData := models.PriceAudit{
		PriceID:         up.ID,
		AssetID:         up.AssetID,
		AggregatedPrice: up,
		RawPrices:       raws,
		CreatedAt:       up.Timestamp,
		UpdatedAt:       up.Timestamp,
	}

	return &auditData, nil
}
