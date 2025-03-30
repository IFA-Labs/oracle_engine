package timescale

import (
	"context"
	"database/sql"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"

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
	// Create prices table if it doesn't exist
	query := `
        CREATE TABLE IF NOT EXISTS prices (
            asset_id TEXT NOT NULL,
            value INT NOT NULL,
            expo SMALLINT NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL,
            source TEXT NOT NULL,
            req_hash TEXT,
            PRIMARY KEY (asset_id, timestamp)
        );
        -- Convert to hypertable for time-series optimization
        SELECT create_hypertable('prices', 'timestamp', if_not_exists => true);
    `
	_, err := t.db.ExecContext(ctx, query)
	if err != nil {
		logging.Logger.Error("Failed to initialize prices table", zap.Error(err))
		return err
	}
	logging.Logger.Info("Prices table initialized")
	return nil
}

func (t *TimescaleDB) SavePrice(ctx context.Context, price models.UnifiedPrice) error {
	query := `
        INSERT INTO prices (asset_id, value, expo, timestamp, source, req_hash)
        VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := t.db.ExecContext(ctx, query,
		price.AssetID, price.Value, price.Expo, price.Timestamp, price.Source, price.ReqHash)
	return err
}

func (t *TimescaleDB) GetLastPrice(ctx context.Context, assetID string) (float64, error) {
	query := `
        SELECT value, expo
        FROM prices
        WHERE asset_id = $1
        ORDER BY timestamp DESC
        LIMIT 1`
	var value int64
	var expo int8
	err := t.db.QueryRowContext(ctx, query, assetID).Scan(&value, &expo)
	if err != nil {
		return 0, err
	}
	logging.Logger.Warn("caught", zap.Int64("key", value), zap.Int32("expo", int32(expo)))
	return models.UnifiedPrice{Value: value, Expo: expo}.Number(), nil
}
