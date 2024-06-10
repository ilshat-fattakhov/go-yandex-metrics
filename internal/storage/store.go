package storage

import (
	"errors"
	"fmt"
	"net/http"

	"go-yandex-metrics/internal/config"
)

type Storage interface {
	SaveMetric(mType, mName, mValue string, w http.ResponseWriter)
	GetMetric(mType, mName string) (string, error)
	GetAllMetrics() string
}

func NewStore(cfg *config.ServerCfg) (Storage, error) {
	switch {
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
		return nil, errors.New("error creating storage")
	}
}
