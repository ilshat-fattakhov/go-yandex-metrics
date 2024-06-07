package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

type FileStorage struct {
	MemStore *MemStorage `json:"data"`
	savePath string
}

func NewFileStorage(cfg config.ServerCfg) (Storage, error) {
	memStore, err := NewMemStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating memory storage: %v", err)
	}

	if cfg.StorageCfg.Restore {
		err := LoadMetrics(memStore, cfg.StorageCfg.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("got error loading metrics from file: %v", err)
		}
		return &FileStorage{
			MemStore: memStore,
			savePath: cfg.StorageCfg.FileStoragePath,
		}, nil
	} else {
		return &FileStorage{
			MemStore: memStore,
			savePath: cfg.StorageCfg.FileStoragePath,
		}, nil
	}
}

func LoadMetrics(s *MemStorage, filePath string) error {
	lg := logger.InitLogger()

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
		fmt.Println(data)
		if err != nil {
			return fmt.Errorf("cannot read storage file: %w", err)
		}

		if err := json.Unmarshal(data, &s); err != nil {
			// file is empty so far, return memory storage ???
			return fmt.Errorf("cannot unmarshal storage file, file is probably empty: %v", err)
		}
		fmt.Println(s)
		lg.Info("finished loading metrics")
		return nil
	}
}

func SaveMetrics(s Storage, filePath string) error {
	lg := logger.InitLogger()

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

func (f *FileStorage) SaveMetric(mType, mName, mValue string, w http.ResponseWriter) {
	f.MemStore.SaveMetric(mType, mName, mValue, w)
}

func (f *FileStorage) GetMetric(mType, mName string) (string, error) {
	return f.MemStore.GetMetric(mType, mName)
}

func (f *FileStorage) GetAllMetrics() string {
	return f.MemStore.GetAllMetrics()
}
