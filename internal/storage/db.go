package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"

	"go.uber.org/zap"
)

type DBStorage struct {
	db *sql.DB
}

const CreateGaugeTableSQL = `
CREATE TABLE IF NOT EXISTS gaugemetrics (
	id SERIAL PRIMARY KEY,
	name VARCHAR (25) UNIQUE NOT NULL,
	value DOUBLE PRECISION NOT NULL
)`

const CreateCounterTableSQL = `
CREATE TABLE IF NOT EXISTS countermetrics (
	id SERIAL PRIMARY KEY,
	name VARCHAR (25) UNIQUE NOT NULL,
	value INTEGER NOT NULL
)`

func NewDBStorage(cfg *config.ServerCfg) (*DBStorage, error) {
	db, err := sql.Open("pgx", cfg.StorageCfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to reach database: %w", err)
	}

	_, err = db.Exec(CreateGaugeTableSQL)
	if err != nil {
		return nil, fmt.Errorf("cannot create gauge metrics table: %w", err)
	}

	_, err = db.Exec(CreateCounterTableSQL)
	if err != nil {
		return nil, fmt.Errorf("cannot create counter metrics table: %w", err)
	}

	return &DBStorage{
		db: db,
	}, nil
}

func LoadMetricsDB(f *DBStorage, filePath string) error {
	return nil
}

func SaveMetricsDB(s Storage, filePath string) error {
	return nil
}

func (d *DBStorage) SaveMetric(mType, mName, mValue string) error {
	lg, err := logger.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	sqlInsert := ""

	switch mType {
	case CounterType:
		sqlInsert = "INSERT INTO countermetrics (name, value) VALUES ($1, $2)" +
			"ON CONFLICT (name) DO UPDATE SET value = $3"
	case GaugeType:
		sqlInsert = "INSERT INTO gaugemetrics (name, value) VALUES ($1, $2)" +
			"ON CONFLICT (name) DO UPDATE SET value = $3"
	}

	ctx := context.Background()
	stmt, err := d.db.PrepareContext(ctx, sqlInsert)
	if err != nil {
		return fmt.Errorf("cannot prepare SQL statement: %w", err)
	}

	_, err = stmt.ExecContext(ctx, mName, mValue, mValue)
	if err != nil {
		return fmt.Errorf("cannot execute sql: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			lg.Error("cannot close statement: %w", zap.Error(err))
		}
	}()

	return nil
}

func (d *DBStorage) GetMetric(mType, mName string) (string, error) {
	lg, err := logger.InitLogger()
	if err != nil {
		return "", fmt.Errorf("failed to init logger: %w", err)
	}

	sqlSelect := ""

	switch mType {
	case CounterType:
		sqlSelect = "SELECT value FROM countermetrics WHERE name=$1"
	case GaugeType:
		sqlSelect = "SELECT value FROM gaugemetrics WHERE name=$1"
	}

	ctx := context.Background()
	stmt, err := d.db.PrepareContext(ctx, sqlSelect)
	if err != nil {
		return "", fmt.Errorf("cannot prepare SQL statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			lg.Error("cannot close statement: %w", zap.Error(err))
		}
	}()

	row := stmt.QueryRowContext(ctx, mName)
	if mType == GaugeType {
		var value float64
		err := row.Scan(&value)
		if err != nil {
			return "", fmt.Errorf("cannot get gauge metric: %w", err)
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	} else {
		var value int64
		err := row.Scan(&value)
		if err != nil {
			return "", fmt.Errorf("cannot get counter metric: %w", err)
		}
		return strconv.FormatInt(value, 10), nil
	}
}

func (d *DBStorage) GetAllMetrics() string {
	return ""
}
