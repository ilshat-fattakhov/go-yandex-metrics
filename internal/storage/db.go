package storage

import (
	"fmt"

	"go-yandex-metrics/internal/config"
)

type DBStorage struct {
	MemStore    *MemStorage `json:"data"`
	databaseDSN string
}

func NewDBStorage(cfg *config.ServerCfg) (*DBStorage, error) {
	MemStore, err := NewMemStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating memory storage: %w", err)
	}

	DBStorage := &DBStorage{
		MemStore:    MemStore,
		databaseDSN: cfg.StorageCfg.DatabaseDSN,
	}

	return DBStorage, nil
}

func LoadMetricsDB(f *DBStorage, filePath string) error {
	return nil
}

func SaveMetricsDB(s Storage, filePath string) error {
	return nil
}

func (d *DBStorage) SaveMetric(mType, mName, mValue string) error {
	if err := d.MemStore.SaveMetric(mType, mName, mValue); err != nil {
		return fmt.Errorf("cannot save metric: %w", err)
	}
	return nil
}

func (d *DBStorage) GetMetric(mType, mName string) (string, error) {
	val, err := d.MemStore.GetMetric(mType, mName)
	if err != nil {
		return "", fmt.Errorf("cannot get metric: %w", err)
	}
	return val, nil
}

func (d *DBStorage) GetAllMetrics() string {
	return d.MemStore.GetAllMetrics()
}
