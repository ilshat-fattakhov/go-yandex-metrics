package storage

import (
	"strconv"
)

// данная структура хранит метрики

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (m *MemStorage) Save(t, n, v string) error {

	if t == "counter" {

		// в случае если мы по какой-то причине получили число с плавающей точкой
		vFloat64, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		vInt64 := int64(vFloat64)
		// новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу
		m.counter[n] += vInt64
		return nil

	} else if t == "gauge" {

		vFloat64, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		// новое значение должно замещать предыдущее
		m.gauge[n] = vFloat64
		return nil

	}
	return nil
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
