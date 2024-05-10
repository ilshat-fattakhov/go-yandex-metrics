package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"go-yandex-metrics/cmd/config"
	"go-yandex-metrics/internal/storage"
)

func RunAgent() error {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	pollInterval, err := strconv.ParseUint(cfg.Agent.PollInterval, 10, 64)
	if err != nil {
		log.Fatal(fmt.Printf("failed to parse %s as a poll interval value: %v", cfg.Agent.PollInterval, err))
	}
	tickerSave := time.NewTicker(time.Duration(pollInterval) * time.Second)

	reportInterval, err := strconv.ParseUint(cfg.Agent.ReportInterval, 10, 64)
	if err != nil {
		log.Fatal(fmt.Printf("failed to parse %s as a report interval value: %v", cfg.Agent.ReportInterval, err))
	}
	tickerSend := time.NewTicker(time.Duration(reportInterval) * time.Second)

	var metricsToSend = storage.NewMemStorage()

	for {
		select {
		case <-tickerSave.C:
			SaveMetrics(metricsToSend)
		case <-tickerSend.C:
			SendMetrics(metricsToSend, cfg.Agent.Host)
		}
	}
}

func SaveMetrics(metricsToSend *storage.MemStorage) {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	metricsToSend.Gauge["Alloc"] = float64(m.Alloc)
	metricsToSend.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	metricsToSend.Gauge["Frees"] = float64(m.Frees)
	metricsToSend.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	metricsToSend.Gauge["GCSys"] = float64(m.GCSys)
	metricsToSend.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	metricsToSend.Gauge["HeapIdle"] = float64(m.HeapIdle)
	metricsToSend.Gauge["HeapInuse"] = float64(m.HeapInuse)
	metricsToSend.Gauge["HeapObjects"] = float64(m.HeapObjects)
	metricsToSend.Gauge["HeapReleased"] = float64(m.HeapReleased)
	metricsToSend.Gauge["HeapSys"] = float64(m.HeapSys)
	metricsToSend.Gauge["LastGC"] = float64(m.LastGC)
	metricsToSend.Gauge["Lookups"] = float64(m.Lookups)
	metricsToSend.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	metricsToSend.Gauge["MCacheSys"] = float64(m.MCacheSys)
	metricsToSend.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	metricsToSend.Gauge["MSpanSys"] = float64(m.MSpanSys)
	metricsToSend.Gauge["Mallocs"] = float64(m.Mallocs)
	metricsToSend.Gauge["NextGC"] = float64(m.NextGC)
	metricsToSend.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	metricsToSend.Gauge["NumGC"] = float64(m.NumGC)
	metricsToSend.Gauge["OtherSys"] = float64(m.OtherSys)
	metricsToSend.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	metricsToSend.Gauge["StackInuse"] = float64(m.StackInuse)
	metricsToSend.Gauge["StackSys"] = float64(m.StackSys)
	metricsToSend.Gauge["Sys"] = float64(m.Sys)
	metricsToSend.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
	metricsToSend.Gauge["RandomValue"] = rand.Float64()
	metricsToSend.Counter["PollCount"]++
}

func SendMetrics(metricsToSend *storage.MemStorage, host string) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	// var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	for n, v := range metricsToSend.Gauge {
		value := strconv.FormatFloat(v, 'f', -1, 64)

		base := "http://" + host
		path := "/update/gauge/" + n + "/" + value
		sendURL, err := url.JoinPath(base, path)
		// logger.Info("Sending gauge metrics to " + sendUrl)

		if err != nil {
			log.Fatal(fmt.Printf("failed to join path parts: %v", err))
		}

		req, err := http.NewRequest(http.MethodPost, sendURL, http.NoBody)
		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		if err != nil {
			continue
		}

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(fmt.Printf("failed to do a request: %v", err))
		}

		if resp.StatusCode != http.StatusOK {
			log.Fatal(fmt.Printf("unexpected response code while sending gauge metrics: %v", resp.StatusCode))
		}

		err = resp.Body.Close()
		if err != nil {
			log.Fatal(fmt.Printf("failed to close response body: %v", err))
		}
	}

	for n, v := range metricsToSend.Counter {
		value := strconv.Itoa(int(v))
		base := "http://" + host
		path := "/update/counter/" + n + "/" + value
		sendURL, err := url.JoinPath(base, path)

		// logger.Info("Sending counter metrics to " + sendUrl)
		if err != nil {
			log.Fatal(fmt.Printf("failed to join path parts: %v", err))
		}
		req, err := http.NewRequest(http.MethodPost, sendURL, http.NoBody)
		if err != nil {
			continue
		}

		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(fmt.Printf("failed to do a request: %v", err))
		}

		if resp.StatusCode != http.StatusOK {
			log.Fatal(fmt.Printf("unexpected response code while sending counter metrics: %v", resp.StatusCode))
		}

		err = resp.Body.Close()
		if err != nil {
			log.Fatal(fmt.Printf("failed to close response body: %v", err))
		}

		// clear metricsToSend PollCount here
		metricsToSend.Counter["PollCount"] = 0
	}
}
