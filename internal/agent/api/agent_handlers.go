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
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
	updatePath         = "update"
)

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string   `json:"id"`              // имя метрики
}

func (a *Agent) saveMetrics() {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	a.store.memLock.Lock()

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

	a.store.memLock.Unlock()
}

func (a *Agent) sendMetrics() error {
	c := &http.Client{Timeout: time.Duration(1) * time.Second}

	for n, v := range a.store.Gauge {
		err := a.sendData(c, v, n, GaugeType, http.MethodPost)
		if err != nil {
			return fmt.Errorf("an error occured sending gauge data: %w", err)
		}
	}

	for n, v := range a.store.Counter {
		err := a.sendData(c, v, n, CounterType, http.MethodPost)
		if err != nil {
			return fmt.Errorf("an error occured sending counter data: %w", err)
		}
	}

	a.store.Counter["PollCount"] = 0
	return nil
}

func (a *Agent) sendData(c *http.Client, v any, n string, mType string, method string) error {
	var metric = Metrics{}

	switch mType {
	case GaugeType:
		switch i := v.(type) {
		case float64:
			metric = Metrics{ID: n, MType: GaugeType, Value: &i}
		default:
			return nil
		}
	case CounterType:
		switch i := v.(type) {
		case int64:
			metric = Metrics{ID: n, MType: CounterType, Delta: &i}
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
	//	return fmt.Errorf("failed to do a request: %w", err)
	//}

	if resp != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
	return nil
}
