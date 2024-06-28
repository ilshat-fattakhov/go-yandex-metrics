package storage

import (
	"fmt"

	"go-yandex-metrics/internal/config"
)

type Storage interface {
	SaveMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (string, error)
	GetAllMetrics() (string, error)
}

func NewStore(cfg *config.ServerCfg) (Storage, error) {
	switch {
	case cfg.StorageCfg.DatabaseDSN != "":
		store, err := NewDBStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating db storage: %w", err)
		}
		return store, nil
	case cfg.StorageCfg.FileStoragePath != "":
		store, err := NewFileStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating file storage: %w", err)
		}
		return store, nil
	case cfg.StorageCfg.FileStoragePath == "":
		store, err := NewMemStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating memory storage: %w", err)
		}
		return store, nil
	default:
		store, err := NewMemStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating memory storage: %w", err)
		}
		return store, nil
	}
}
