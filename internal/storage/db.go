package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go-yandex-metrics/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStorage struct {
	pool *pgxpool.Pool
}

type Metric struct {
	memLock     *sync.Mutex
	metricName  string
	metricValue string
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func NewDBStorage(cfg *config.ServerCfg) (*DBStorage, error) {
	if err := runMigrations(cfg.StorageCfg.DatabaseDSN); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.StorageCfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	return &DBStorage{
		pool: pool,
	}, nil
}

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return iofs driver: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (d *DBStorage) SaveMetric(mType, mName, mValue string) error {
	sqlInsert := ""

	switch mType {
	case CounterType:
		sqlInsert = "INSERT INTO countermetrics (metricName, metricValue) VALUES ($1, $2)" +
			"ON CONFLICT (metricName) DO UPDATE SET metricValue = countermetrics.metricValue + $3"
	case GaugeType:
		sqlInsert = "INSERT INTO gaugemetrics (metricName, metricValue) VALUES ($1, $2)" +
			"ON CONFLICT (metricName) DO UPDATE SET metricValue = $3"
	}

	ctx := context.Background()
	_, err := d.pool.Exec(ctx, sqlInsert, mName, mValue, mValue)
	if err != nil {
		return fmt.Errorf("cannot execute query while saving metric: %w", err)
	}

	return nil
}

func (d *DBStorage) GetMetric(mType, mName string) (string, error) {
	sqlSelect := ""

	switch mType {
	case CounterType:
		sqlSelect = "SELECT metricValue FROM countermetrics WHERE metricName=$1"
	case GaugeType:
		sqlSelect = "SELECT metricValue FROM gaugemetrics WHERE metricName=$1"
	}

	ctx := context.Background()
	row := d.pool.QueryRow(ctx, sqlSelect, mName)

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
	html := "<h3>Gauge:</h3>"
	gaugeMetrics, err := d.getMetrics(GaugeType)
	if err != nil {
		return "", fmt.Errorf("error getting gauge metrics: %w", err)
	}
	html += gaugeMetrics

	html += "<h3>Counter:</h3>"
	counterMetrics, err := d.getMetrics(CounterType)
	if err != nil {
		return "", fmt.Errorf("error getting counter metrics: %w", err)
	}
	html += counterMetrics

	return html, nil
}

func (d *DBStorage) getMetrics(mType string) (string, error) {
	sqlSelect := ""

	switch mType {
	case CounterType:
		sqlSelect = "SELECT metricName, metricValue FROM countermetrics"
	case GaugeType:
		sqlSelect = "SELECT metricName, metricValue FROM gaugemetrics"
	}

	ctx := context.Background()
	rows, err := d.pool.Query(ctx, sqlSelect)
	if err != nil {
		return "", fmt.Errorf("error running sql query: %w", err)
	}

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
		m.memLock.Lock()
		metrics = append(metrics, m)
		m.memLock.Unlock()
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
