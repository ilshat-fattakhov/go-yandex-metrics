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

const (
	CreateGaugeTableSQL = `
CREATE TABLE IF NOT EXISTS gaugemetrics (
	id SERIAL PRIMARY KEY,
	metricName VARCHAR (25) UNIQUE NOT NULL,
	metricValue DOUBLE PRECISION NOT NULL
)`

	CreateCounterTableSQL = `
CREATE TABLE IF NOT EXISTS countermetrics (
	id SERIAL PRIMARY KEY,
	metricName VARCHAR (25) UNIQUE NOT NULL,
	metricValue INTEGER NOT NULL
)`
	failedToInitLogger = "failed to init logger"
)

type Metric struct {
	metricName  string
	metricValue string
}

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
		return fmt.Errorf(failedToInitLogger+": %w", err)
	}
	sqlInsert := ""

	switch mType {
	case CounterType:
		sqlInsert = "INSERT INTO countermetrics (metricName, metricValue) VALUES ($1, $2)" +
			"ON CONFLICT (metricName) DO UPDATE SET metricValue = $3"
	case GaugeType:
		sqlInsert = "INSERT INTO gaugemetrics (metricName, metricValue) VALUES ($1, $2)" +
			"ON CONFLICT (metricName) DO UPDATE SET metricValue = $3"
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
		return "", fmt.Errorf(failedToInitLogger+": %w", err)
	}

	sqlSelect := ""

	switch mType {
	case CounterType:
		sqlSelect = "SELECT metricValue FROM countermetrics WHERE metricName=$1"
	case GaugeType:
		sqlSelect = "SELECT metricValue FROM gaugemetrics WHERE metricName=$1"
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
		var metricValue float64
		err := row.Scan(&metricValue)
		if err != nil {
			return "", fmt.Errorf("cannot get gauge metric: %w", err)
		}
		return strconv.FormatFloat(metricValue, 'f', -1, 64), nil
	} else {
		var metricValue int64
		err := row.Scan(&metricValue)
		if err != nil {
			return "", fmt.Errorf("cannot get counter metric: %w", err)
		}
		return strconv.FormatInt(metricValue, 10), nil
	}
}

func (d *DBStorage) GetAllMetrics() (string, error) {
	html := "<h3>Gaugesa:</h3>"
	gaugeMetrics, err := d.getMetrics(GaugeType)
	if err != nil {
		return "", fmt.Errorf("error getting gauge metrics: %w", err)
	}
	html += gaugeMetrics

	html += "<h3>Countersa:</h3>"
	counterMetrics, err := d.getMetrics(CounterType)
	if err != nil {
		return "", fmt.Errorf("error getting counter metrics: %w", err)
	}
	html += counterMetrics

	return html, nil
}

func (d *DBStorage) getMetrics(mType string) (string, error) {
	lg, err := logger.InitLogger()
	if err != nil {
		return "", fmt.Errorf(failedToInitLogger+": %w", err)
	}
	sqlSelect := ""

	switch mType {
	case CounterType:
		sqlSelect = "SELECT metricName, metricValue FROM countermetrics"
	case GaugeType:
		sqlSelect = "SELECT metricName, metricValue FROM gaugemetrics"
	}

	ctx := context.Background()
	rows, err := d.db.QueryContext(ctx, sqlSelect)
	if err != nil {
		return "", fmt.Errorf("error running sql query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			lg.Error("error running sql query: %w", zap.Error(err))
		}
	}()

	metrics := make([]Metric, 0)

	for rows.Next() {
		var m Metric
		if err != nil {
			return "", fmt.Errorf("cannot scan row: %w", err)
		}
		if mType == GaugeType {
			err = rows.Scan(&m.metricName, &m.metricValue)
			if err != nil {
				return "", fmt.Errorf("cannot get gauge metric: %w", err)
			}
		} else {
			err = rows.Scan(&m.metricName, &m.metricValue)
			if err != nil {
				return "", fmt.Errorf("cannot get counter metric: %w", err)
			}
		}
		metrics = append(metrics, m)
	}

	err = rows.Err()
	if err != nil {
		return "", fmt.Errorf("error fetching rows from the db: %w", err)
	}
	html := ""
	for _, v := range metrics {
		html += v.metricName + " : " + v.metricValue + "<br>"
	}

	return html, nil
}
