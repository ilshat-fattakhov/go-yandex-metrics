package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
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
	memLock sync.Mutex
}

type FileStorage struct {
	MemStore *MemStorage `json:"data"`
	SavePath string      `json:"file_storage_path"`
}

type Store struct {
	MemStore *MemStorage `json:"data"`
	// FileStore *FileStorage
	// db     *sql.DB
}

type StorageCfg struct {
	FileStoragePath string `json:"file_storage_path"`
	StoreInterval   uint64 `json:"store_interval"`
	Restore         bool   `json:"restore"`
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

func NewStore(cfg config.ServerCfg) (*Store, error) {
	lg := logger.InitLogger()

	var store = &Store{
		MemStore: NewMemStorage(),
		// FileStore: NewFileStorage(),
	}

	if cfg.StorageCfg.FileStoragePath != "" {
		if cfg.StorageCfg.Restore {
			_, err := Load(store, cfg.StorageCfg.FileStoragePath)
			if err != nil {
				lg.Info("got error loading metrics from file: " + cfg.StorageCfg.FileStoragePath)
				return nil, err
			}
		}
	}

	return store, nil
}

func (s *Store) Save(filePath string) error {
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

func Load(s *Store, filePath string) (*Store, error) {
	lg := logger.InitLogger()

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				lg.Info(fmt.Sprintf("error creating storage file: %v", err))
				return nil, fmt.Errorf("error creating storage file: %w", err)
			}
			return &Store{
				MemStore: NewMemStorage(),
				// FileStore: NewFileStorage(),
			}, nil
		} else {
			lg.Info(fmt.Sprintf("storage file error: %v", err))
			return nil, fmt.Errorf("storage file error: %w", err)
		}
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			lg.Info(fmt.Sprintf("cannot read storage file: %v", err))
			return nil, fmt.Errorf("cannot read storage file: %w", err)
		}
		if err := json.Unmarshal(data, s); err != nil {
			// file is empty so far, return memory storage
			lg.Info(fmt.Sprintf("cannot unmarshal storage file: %v", err))
			return &Store{
				MemStore: NewMemStorage(),
				// FileStore: NewFileStorage(),
			}, nil
		}
		return s, nil
	}
}

func SaveMetric(store *Store, mType, mName, mValue string, w http.ResponseWriter) {
	switch mType {
	case CounterType:
		saveCounter(store, mName, mValue, w)
	case GaugeType:
		saveGauge(store, mName, mValue, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func saveCounter(store *Store, mName, mValue string, w http.ResponseWriter) {
	lg := logger.InitLogger()

	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		lg.Info(fmt.Sprintf("got error pasring float value for counter metric: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	store.MemStore.memLock.Lock()
	store.MemStore.Counter[mName] += vInt64
	store.MemStore.memLock.Unlock()
}

func saveGauge(store *Store, mName, mValue string, w http.ResponseWriter) {
	lg := logger.InitLogger()

	vFloat64, err := strconv.ParseFloat(mValue, 64)

	if err != nil {
		lg.Info("Got error pasring float value for gauge metric")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	store.MemStore.memLock.Lock()
	store.MemStore.Gauge[mName] = vFloat64
	store.MemStore.memLock.Unlock()
}
