package hadnlers

import (
	"go-yandex-metrics/internal/storage"
	"math/rand"
	"runtime"
)

func SaveMetrics(m *runtime.MemStats) {

	//logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	//logger.Info("Saving metrics...")

	storage.GaugeMetrics["Alloc"] = float64(m.Alloc)
	storage.GaugeMetrics["BuckHashSys"] = float64(m.BuckHashSys)
	storage.GaugeMetrics["Frees"] = float64(m.Frees)
	storage.GaugeMetrics["GCCPUFraction"] = float64(m.GCCPUFraction)
	storage.GaugeMetrics["GCSys"] = float64(m.GCSys)
	storage.GaugeMetrics["HeapAlloc"] = float64(m.HeapAlloc)
	storage.GaugeMetrics["HeapIdle"] = float64(m.HeapIdle)
	storage.GaugeMetrics["HeapInuse"] = float64(m.HeapInuse)
	storage.GaugeMetrics["HeapObjects"] = float64(m.HeapObjects)
	storage.GaugeMetrics["HeapReleased"] = float64(m.HeapReleased)
	storage.GaugeMetrics["HeapSys"] = float64(m.HeapSys)
	storage.GaugeMetrics["LastGC"] = float64(m.LastGC)
	storage.GaugeMetrics["Lookups"] = float64(m.Lookups)
	storage.GaugeMetrics["MCacheInuse"] = float64(m.MCacheInuse)
	storage.GaugeMetrics["MCacheSys"] = float64(m.MCacheSys)
	storage.GaugeMetrics["MSpanInuse"] = float64(m.MSpanInuse)
	storage.GaugeMetrics["MSpanSys"] = float64(m.MSpanSys)
	storage.GaugeMetrics["Mallocs"] = float64(m.Mallocs)
	storage.GaugeMetrics["NextGC"] = float64(m.NextGC)
	storage.GaugeMetrics["NumForcedGC"] = float64(m.NumForcedGC)
	storage.GaugeMetrics["NumGC"] = float64(m.NumGC)
	storage.GaugeMetrics["OtherSys"] = float64(m.OtherSys)
	storage.GaugeMetrics["PauseTotalNs"] = float64(m.PauseTotalNs)
	storage.GaugeMetrics["StackInuse"] = float64(m.StackInuse)
	storage.GaugeMetrics["StackSys"] = float64(m.StackSys)
	storage.GaugeMetrics["Sys"] = float64(m.Sys)
	storage.GaugeMetrics["TotalAlloc"] = float64(m.TotalAlloc)
	storage.GaugeMetrics["RandomValue"] = rand.Float64() * 5

	//logger.Info("Saved metrics...")

}
