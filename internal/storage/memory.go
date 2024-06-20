package storage

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"go-yandex-metrics/internal/config"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
	memLock *sync.Mutex
}

func NewMemStorage(cfg *config.ServerCfg) (*MemStorage, error) {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
		memLock: &sync.Mutex{},
	}, nil
}

func (m *MemStorage) SaveMetric(mType, mName, mValue string) error {
	switch mType {
	case CounterType:
		if err := m.saveCounter(mName, mValue); err != nil {
			return fmt.Errorf("failed to save counter: %w", err)
		}
	case GaugeType:
		if err := m.saveGauge(mName, mValue); err != nil {
			return fmt.Errorf("failed to save gauge: %w", err)
		}
	default:
		return fmt.Errorf("wrong metric type: %v", mType)
	}
	return nil
}

func (m *MemStorage) GetMetric(mType, mName string) (string, error) {
	var html string

	switch mType {
	case GaugeType:
		if mValue, ok := m.Gauge[mName]; !ok {
			return "", errors.New(GaugeType + " metric not found")
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case CounterType:
		if mValue, ok := m.Counter[mName]; !ok {
			return "", errors.New(CounterType + " metric not found")
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}
}

func (m *MemStorage) GetAllMetrics() string {
	html := "<h3>Gauge:</h3>"
	for mName, mValue := range m.Gauge {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range m.Counter {
		html += (mName + ":" + strconv.FormatInt(mValue, 10) + "<br>")
	}
	return html
}

func (m *MemStorage) saveCounter(mName, mValue string) error {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return fmt.Errorf("got error parsing float value for counter metric: %w", err)
	}

	vInt64 := int64(vFloat64)
	m.memLock.Lock()
	m.Counter[mName] += vInt64
	m.memLock.Unlock()

	return nil
}

func (m *MemStorage) saveGauge(mName, mValue string) error {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		return fmt.Errorf("got error parsing float value for gauge metric: %w", err)
	}

	m.memLock.Lock()
	m.Gauge[mName] = vFloat64
	m.memLock.Unlock()
	return nil
}
