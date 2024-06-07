package storage

import (
	"errors"
	"fmt"
	"net/http"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

type Storage interface {
	SaveMetric(mType, mName, mValue string, w http.ResponseWriter)
	GetMetric(mType, mName string) (string, error)
	GetAllMetrics() string
}

func NewStore(cfg config.ServerCfg) (Storage, error) {
	lg := logger.InitLogger()

	switch {
	case cfg.StorageCfg.FileStoragePath != "":
		store, err := NewFileStorage(cfg)
		if err != nil {
			lg.Info(fmt.Sprintf("error creating file storage: %v", err))
			return nil, errors.New("error creating file storage")
		}
		return store, nil
	case cfg.StorageCfg.FileStoragePath == "":
		store, err := NewMemStorage(cfg)
		if err != nil {
			lg.Info(fmt.Sprintf("error creating memory storage: %v", err))
			return nil, errors.New("error creating memory storage")
		}
		return store, nil
	default:
		lg.Info("error creating storage")
		return nil, errors.New("error creating storage")
	}
}
