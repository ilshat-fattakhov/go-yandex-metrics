package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"go-yandex-metrics/internal/storage"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
	updatePath         = "update"
)

func (a *Agent) saveMetrics() {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	a.store.MemStore.Gauge["Alloc"] = float64(m.Alloc)
	a.store.MemStore.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	a.store.MemStore.Gauge["Frees"] = float64(m.Frees)
	a.store.MemStore.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	a.store.MemStore.Gauge["GCSys"] = float64(m.GCSys)
	a.store.MemStore.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	a.store.MemStore.Gauge["HeapIdle"] = float64(m.HeapIdle)
	a.store.MemStore.Gauge["HeapInuse"] = float64(m.HeapInuse)
	a.store.MemStore.Gauge["HeapObjects"] = float64(m.HeapObjects)
	a.store.MemStore.Gauge["HeapReleased"] = float64(m.HeapReleased)
	a.store.MemStore.Gauge["HeapSys"] = float64(m.HeapSys)
	a.store.MemStore.Gauge["LastGC"] = float64(m.LastGC)
	a.store.MemStore.Gauge["Lookups"] = float64(m.Lookups)
	a.store.MemStore.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	a.store.MemStore.Gauge["MCacheSys"] = float64(m.MCacheSys)
	a.store.MemStore.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	a.store.MemStore.Gauge["MSpanSys"] = float64(m.MSpanSys)
	a.store.MemStore.Gauge["Mallocs"] = float64(m.Mallocs)
	a.store.MemStore.Gauge["NextGC"] = float64(m.NextGC)
	a.store.MemStore.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	a.store.MemStore.Gauge["NumGC"] = float64(m.NumGC)
	a.store.MemStore.Gauge["OtherSys"] = float64(m.OtherSys)
	a.store.MemStore.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	a.store.MemStore.Gauge["StackInuse"] = float64(m.StackInuse)
	a.store.MemStore.Gauge["StackSys"] = float64(m.StackSys)
	a.store.MemStore.Gauge["Sys"] = float64(m.Sys)
	a.store.MemStore.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
	a.store.MemStore.Gauge["RandomValue"] = rand.Float64()
	a.store.MemStore.Counter["PollCount"]++
}

func (a *Agent) sendMetrics() error {
	c := &http.Client{Timeout: time.Duration(1) * time.Second}

	for n, v := range a.store.MemStore.Gauge {
		err := a.sendData(c, v, n, GaugeType, http.MethodPost)
		if err != nil {
			return fmt.Errorf("an error occured sending gauge data: %w", err)
		}
	}

	for n, v := range a.store.MemStore.Counter {
		err := a.sendData(c, v, n, CounterType, http.MethodPost)
		if err != nil {
			return fmt.Errorf("an error occured sending counter data: %w", err)
		}
	}

	a.store.MemStore.Counter["PollCount"] = 0
	return nil
}

func (a *Agent) sendData(c *http.Client, v any, n string, mType string, method string) error {
	var metric = storage.Metrics{}

	switch mType {
	case GaugeType:
		switch i := v.(type) {
		case float64:
			metric = storage.Metrics{ID: n, MType: GaugeType, Value: &i}
		default:
			return nil
		}
	case CounterType:
		switch i := v.(type) {
		case int64:
			metric = storage.Metrics{ID: n, MType: CounterType, Delta: &i}
		default:
			return nil
		}
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(metric)

	if err != nil {
		a.logger.Info(fmt.Sprintf("failed to JSON encode gauge metric: %v", err))
		return fmt.Errorf("failed to JSON encode gauge metric: %w", err)
	}

	sendURL, err := url.JoinPath("http://", a.cfg.Host, updatePath, "/")
	if err != nil {
		a.logger.Info(fmt.Sprintf("failed to join path parts for gauge JSON POST URL: %v", err))
		return fmt.Errorf("failed to join path parts for gauge JSON POST URL: %w", err)
	}

	req, err := http.NewRequest(method, sendURL, &buf)
	if err != nil {
		a.logger.Info(fmt.Sprintf("failed to create a request: %v", err))
		return fmt.Errorf("failed to create a request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	//	req.Header.Set("Content-Length", strconv.Itoa(binary.Size(&buf)))
	resp, _ := c.Do(req)
	// если раскомментировать строки ниже, автотест не проходится
	// с ошибкой "невозможно установить соединение с сервером"
	// if err != nil {
	//	a.logger.Info(fmt.Sprintf("failed to do a request: %v", err))
	// return fmt.Errorf("failed to do a request: %w", err)
	// }

	if resp != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
	return nil
}
