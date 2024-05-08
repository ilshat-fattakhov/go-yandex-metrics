package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"go-yandex-metrics/internal/storage"
)

func SaveMetrics(m *runtime.MemStats) {
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
	storage.GaugeMetrics["RandomValue"] = rand.Float64()
}

func SendMetrics(host string) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	for n, v := range storage.GaugeMetrics {

		value := fmt.Sprintf("%f", v)
		base := "http://" + host
		path := "/update/gauge/" + n + "/" + value
		url, err := url.JoinPath(base, path)
		if err != nil {
			log.Fatal(fmt.Printf("failed to join path parts: %v", err))
		}

		req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		if err != nil {
			continue
		}

		resp, err := c.Do(req)
		if err != nil {
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			log.Fatal(fmt.Printf("failed to close response body: %v", err))
		}
	}

	for n := range storage.CounterMetrics {
		base := "http://" + host
		path := "/update/counter/" + n + "/1"
		url, err := url.JoinPath(base, path)
		if err != nil {
			log.Fatal(fmt.Printf("failed to join path parts: %v", err))
		}
		req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
		if err != nil {
			continue
		}

		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(fmt.Printf("failed to do a request: %v", err))
		}

		if resp.StatusCode != 200 {
			log.Fatal(fmt.Printf("unexpected response code: %v", resp.StatusCode))
		}

		err = resp.Body.Close()
		if err != nil {
			log.Fatal(fmt.Printf("failed to close response body: %v", err))
		}
	}
}
