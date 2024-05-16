package api

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"go-yandex-metrics/internal/config"
)

func (a *Agent) SaveMetrics() error {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	a.store.Gauge["Alloc"] = float64(m.Alloc)
	a.store.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	a.store.Gauge["Frees"] = float64(m.Frees)
	a.store.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	a.store.Gauge["GCSys"] = float64(m.GCSys)
	a.store.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	a.store.Gauge["HeapIdle"] = float64(m.HeapIdle)
	a.store.Gauge["HeapInuse"] = float64(m.HeapInuse)
	a.store.Gauge["HeapObjects"] = float64(m.HeapObjects)
	a.store.Gauge["HeapReleased"] = float64(m.HeapReleased)
	a.store.Gauge["HeapSys"] = float64(m.HeapSys)
	a.store.Gauge["LastGC"] = float64(m.LastGC)
	a.store.Gauge["Lookups"] = float64(m.Lookups)
	a.store.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	a.store.Gauge["MCacheSys"] = float64(m.MCacheSys)
	a.store.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	a.store.Gauge["MSpanSys"] = float64(m.MSpanSys)
	a.store.Gauge["Mallocs"] = float64(m.Mallocs)
	a.store.Gauge["NextGC"] = float64(m.NextGC)
	a.store.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	a.store.Gauge["NumGC"] = float64(m.NumGC)
	a.store.Gauge["OtherSys"] = float64(m.OtherSys)
	a.store.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	a.store.Gauge["StackInuse"] = float64(m.StackInuse)
	a.store.Gauge["StackSys"] = float64(m.StackSys)
	a.store.Gauge["Sys"] = float64(m.Sys)
	a.store.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
	a.store.Gauge["RandomValue"] = rand.Float64()
	a.store.Counter["PollCount"]++

	return nil
}

func (a *Agent) SendMetrics() error {
	for n, v := range a.store.Gauge {
		value := strconv.FormatFloat(v, 'f', -1, 64)

		sendURL, err := url.JoinPath("http://", a.cfg.Host, "update", config.GaugeType, n, value)
		if err != nil {
			log.Printf("failed to join path parts: %v", err)
			return nil
		}

		sendData(http.MethodPost, sendURL)
	}

	for n, v := range a.store.Counter {
		value := strconv.Itoa(int(v))
		sendURL, err := url.JoinPath("http://", a.cfg.Host, "update", config.CounterType, n, value)
		if err != nil {
			log.Printf("failed to join path parts: %v", err)
			return nil
		}

		sendData(http.MethodPost, sendURL)
	}

	a.store.Counter["PollCount"] = 0
	return nil
}

func sendData(method, sendURL string) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	req, err := http.NewRequest(method, sendURL, http.NoBody)
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Content-Length", "0")

	if err != nil {
		log.Printf("failed to create a request: %v", err)
		return
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("failed to do a request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
			return
		}
	}()
	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected response code while sending gauge metrics: %v", resp.StatusCode)
		return
	}
}
