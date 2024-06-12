package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
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
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("error creating storage file: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("unexpected error while creating storage file: %w", err)
		}
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read storage file: %w", err)
		}

		if err := json.Unmarshal(data, f); err != nil {
			return fmt.Errorf("cannot unmarshal storage file, file is probably empty: %w", err)
		}
		return nil
	}
}

func SaveMetricsDB(s Storage, filePath string) error {
	lg, err := logger.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "   ")
	if err != nil {
		lg.Info("cannot marshal storage")
		return fmt.Errorf("cannot marshal storage: %w", err)
	}

	err = os.WriteFile(filePath, data, 0o600)
	if err != nil {
		lg.Info("cannot save storage to file")
		return fmt.Errorf("cannot save storage to file: %w", err)
	}

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
