package storage

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

var Mem = NewMemStorage()

func (mem *MemStorage) saveCounter(mName, mValue string, w http.ResponseWriter) {

	// в случае если мы по какой-то причине получили число с плавающей точкой
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		// Принимаем метрики только по протоколу HTTP методом POST
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	// новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу
	mem.Counter[mName] += vInt64
}

func (mem *MemStorage) saveGauge(mName, mValue string, w http.ResponseWriter) {

	//_, ok := GaugeMetrics["mName"]
	//if ok {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		// Принимаем метрики только по протоколу HTTP методом POST
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// новое значение должно замещать предыдущее
	mem.Gauge[mName] = vFloat64
}

func (mem *MemStorage) Save(mType, mName, mValue string, w http.ResponseWriter) {

	if mType == "counter" {
		mem.saveCounter(mName, mValue, w)
	} else if mType == "gauge" {
		mem.saveGauge(mName, mValue, w)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (mem *MemStorage) Get(mType, mName string, w http.ResponseWriter) string {

	logger.Info("Got type:" + mType + " and name " + mName)
	fmt.Println(mem)
	fmt.Println(mem.Counter)

	if mType == "counter" {
		mValue, ok := mem.Counter[mName]
		if ok {
			logger.Info("Regurning " + mName + " counter value:" + fmt.Sprint(mValue))
			return fmt.Sprint(mValue)
			//return (GetCounter(mName))
		} else {
			return ""
		}

	} else if mType == "gauge" {
		mValue, ok := mem.Gauge[mName]
		if ok {
			logger.Info("Returning " + mName + " counter value:" + fmt.Sprint(mValue))
			return fmt.Sprint(mValue)
		} else {
			return ""
		}

	} else {
		return ""
	}

}

func getCounterValue(mType, mName string) string {
	logger.Info("getCounterValue func: Type Name: " + mType + mName)

	if mType == "gauge" {
		logger.Info("We are in gauge: " + mName)

		if mValue, ok := GaugeMetrics[mName]; ok {
			return fmt.Sprint(mValue)
		}
	} else if mType == "counter" {
		logger.Info("We are in counter: " + mName)

		if mValue, ok := CounterMetrics[mName]; ok {

			//logger.Info("OK" + string(mValue))
			return fmt.Sprint(mValue)
		} else {

			logger.Info("NOT OK")

		}
	}
	return ""
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
