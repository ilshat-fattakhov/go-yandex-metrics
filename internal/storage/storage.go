package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	logger "go-yandex-metrics/cmd/server/middleware"
)

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string   `json:"id"`              // имя метрики
}

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
	MemLock sync.Mutex         `json:"memlock"`
}

type FileStorage struct {
	MemStore *MemStorage `json:"data"`
	SavePath string      `json:"file_storage_path"`
}

type StorageCfg struct {
	FileStoragePath string `json:"file_storage_path"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func NewFileStorage() *FileStorage {
	memStore := NewMemStorage()
	return &FileStorage{
		memStore,
		"",
	}
}

func (f *FileStorage) Save(filePath string) error {
	lg := logger.InitLogger()

	data, err := json.MarshalIndent(f, "", "   ")
	if err != nil {
		lg.Info("Cannot marshal storage")
		return fmt.Errorf("cannot marshal storage: %w", err)
	}

	err = os.WriteFile(filePath, data, 0o600)
	if err != nil {
		lg.Info("Cannot save storage to file")
		return fmt.Errorf("cannot save storage to file: %w", err)
	}

	return nil
}

func (f *FileStorage) Load(filePath string) (*FileStorage, error) {
	lg := logger.InitLogger()

	data, err := os.ReadFile(filePath)
	if err != nil {
		lg.Info("Cannot read storage file")
		return nil, fmt.Errorf("cannot read storage file: %w", err)
	}
	if err := json.Unmarshal(data, f); err != nil {
		lg.Info("Cannot unmarshal storage file")
		return nil, fmt.Errorf("cannot unmarshal storage file: %w", err)
	}

	return f, nil
}
