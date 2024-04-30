package handlers

import (
	"fmt"
	"go-yandex-metrics/cmd/agent/storage"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const ReportInterval = 10 //Отправлять метрики на сервер с заданной частотой: reportInterval — 10 секунд.

func SendMetrics() {

	c := http.Client{Timeout: time.Duration(1) * time.Second}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// send gauge metrics
	logger.Info("Sending gauge metrics...")

	//for n, v := range storage.GaugeMetrics {

	//	url := "http://localhost:8080/update/gauge/" + n + "/" + fmt.Sprintf("%v", v)

	//	logger.Info("Sending to " + url)

	//	req, err := http.NewRequest("POST", url, nil)
	//	if err != nil {
	//		fmt.Printf("error %s", err)
	//		return
	//	}

	//	req.Header.Add("Content-Type", "text/plain")
	//	req.Header.Add("Content-Length", "0")

	//	resp, err := c.Do(req)
	//	if err != nil {
	//		fmt.Printf("error %s", err)
	//		return
	//	}

	//	defer resp.Body.Close()

	//}

	// send counter metrics
	logger.Info("Sending counter metrics...")

	for n := range storage.CounterMetrics {
		// PollCount (тип counter) — счётчик, увеличивающийся на 1 при каждом обновлении метрики
		// из пакета runtime на каждый pollInterval
		url := "http://localhost:8080/update/counter/" + n + "/1"
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			logger.Info(fmt.Sprintf("error %s", err))
			return
		}

		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		resp, err := c.Do(req)
		if err != nil {
			logger.Info(fmt.Sprintf("error %s", err))
			return
		}

		defer resp.Body.Close()

	}
}
