package storage

import (
	"errors"
	"fmt"
	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
	"net/http"
	"strconv"
	"sync"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
	memLock sync.Mutex
}

func NewMemStorage(cfg config.ServerCfg) (*MemStorage, error) {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}, nil
}

func (m *MemStorage) SaveMetric(mType, mName, mValue string, w http.ResponseWriter) {
	switch mType {
	case CounterType:
		m.saveCounter(mName, mValue, w)
	case GaugeType:
		m.saveGauge(mName, mValue, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *MemStorage) GetMetric(mType, mName string) (string, error) {
	var html string
	var ErrItemNotFound = errors.New("item not found")

	switch mType {
	case GaugeType:
		if mValue, ok := m.Gauge[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case CounterType:
		if mValue, ok := m.Counter[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}
}

func (m *MemStorage) GetAllMetrics() string {
	fmt.Println(m)
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

func (m *MemStorage) saveCounter(mName, mValue string, w http.ResponseWriter) {
	lg := logger.InitLogger()

	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		lg.Info(fmt.Sprintf("got error parsing float value for counter metric: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	m.memLock.Lock()
	m.Counter[mName] += vInt64
	m.memLock.Unlock()
}

func (m *MemStorage) saveGauge(mName, mValue string, w http.ResponseWriter) {
	lg := logger.InitLogger()
	lg.Info("saving gauge")

	vFloat64, err := strconv.ParseFloat(mValue, 64)

	if err != nil {
		lg.Info("got error parsing float value for gauge metric")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.memLock.Lock()
	m.Gauge[mName] = vFloat64
	m.memLock.Unlock()
	fmt.Println("here's our repo: ", m)
}
