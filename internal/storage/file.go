package storage

import (
	"encoding/json"
	"fmt"
	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"

	"net/http"
	"os"
)

type FileStorage struct {
	MemStore *MemStorage `json:"data"`
	SavePath string      `json:"file_storage_path"`
}

func NewFileStorage(cfg config.ServerCfg) (Storage, error) {
	lg := logger.InitLogger()

	memStore, err := NewMemStorage(cfg)
	if err != nil {
		lg.Info(fmt.Sprintf("error creating memory storage: %v", err))
		return nil, err
	}

	if cfg.StorageCfg.Restore {
		fmt.Println("restoring...")
		err := LoadMetrics(memStore, cfg.StorageCfg.FileStoragePath)
		if err != nil {
			lg.Info(fmt.Sprintf("got error loading metrics from file: %v", err))
			return nil, err
		}

		return &FileStorage{
			MemStore: memStore,
			SavePath: cfg.StorageCfg.FileStoragePath,
		}, nil
	} else {
		return &FileStorage{
			MemStore: memStore,
			SavePath: cfg.StorageCfg.FileStoragePath,
		}, nil
	}
}

func LoadMetrics(s *MemStorage, filePath string) error {
	lg := logger.InitLogger()

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				lg.Info(fmt.Sprintf("error creating storage file: %v", err))
				return fmt.Errorf("error creating storage file: %w", err)
			}
			return nil
		} else {
			lg.Info(fmt.Sprintf("unexpected error while creating storage file: %v", err))
			return fmt.Errorf("unexpected error while creating storage file: %w", err)
		}
	} else {
		data, err := os.ReadFile(filePath)
		fmt.Println(data)
		if err != nil {
			lg.Info(fmt.Sprintf("cannot read storage file: %v", err))
			return fmt.Errorf("cannot read storage file: %w", err)
		}

		if err := json.Unmarshal(data, s); err != nil {
			// file is empty so far, return memory storage
			lg.Info(fmt.Sprintf("cannot unmarshal storage file, file is probably empty: %v", err))
			return nil
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
