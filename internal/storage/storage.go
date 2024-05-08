package storage

import (
	"net/http"
	"strconv"
)

type MemStorage struct {
	// sync.Mutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

var Mem = NewMemStorage()

func (mem *MemStorage) saveCounter(mName, mValue string, w http.ResponseWriter) {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	mem.Counter[mName] += vInt64
}

func (mem *MemStorage) saveGauge(mName, mValue string, w http.ResponseWriter) {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mem.Gauge[mName] = vFloat64
}

func (mem *MemStorage) Save(mType, mName, mValue string, w http.ResponseWriter) {
	switch mType {
	case "counter":
		mem.saveCounter(mName, mValue, w)
	case "gauge":
		mem.saveGauge(mName, mValue, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (mem *MemStorage) Get(mType, mName string, w http.ResponseWriter) string {
	switch mType {
	case "counter":
		mValue, ok := mem.Counter[mName]
		if ok {
			return strconv.FormatInt(mValue, 10)
		} else {
			return ""
		}
	case "gauge":
		mValue, ok := mem.Gauge[mName]
		if ok {
			return strconv.FormatFloat(mValue, 'f', -1, 64)
		} else {
			return ""
		}
	default:
		return ""
	}
}

var GaugeMetrics = map[string]float64{
	"Alloc":         0,
	"BuckHashSys":   0,
	"Frees":         0,
	"GCCPUFraction": 0,
	"GCSys":         0,
	"HeapAlloc":     0,
	"HeapIdle":      0,
	"HeapInuse":     0,
	"HeapObjects":   0,
	"HeapReleased":  0,
	"HeapSys":       0,
	"LastGC":        0,
	"Lookups":       0,
	"MCacheInuse":   0,
	"MCacheSys":     0,
	"MSpanInuse":    0,
	"MSpanSys":      0,
	"Mallocs":       0,
	"NextGC":        0,
	"NumForcedGC":   0,
	"NumGC":         0,
	"OtherSys":      0,
	"PauseTotalNs":  0,
	"StackInuse":    0,
	"StackSys":      0,
	"Sys":           0,
	"TotalAlloc":    0,
	"RandomValue":   0,
}
var CounterMetrics = map[string]int64{
	"PollCount": 0,
}
