package main

import (
	"fmt"
	send "go-yandex-metrics/cmd/agent/handlers/send"
	"log"
	"log/slog"
	"strconv"

	save "go-yandex-metrics/cmd/agent/handlers/save"

	"go-yandex-metrics/internal/storage"
	"os"
	"runtime"
	"time"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {

	storage.ParseFlags()

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	pollInterval, err := strconv.ParseUint(storage.PollInterval, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	tickerSend := time.NewTicker(time.Duration(pollInterval) * time.Second)
	fmt.Println("Poll interval is", pollInterval)

	reportInterval, err := strconv.ParseUint(storage.ReportInterval, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	tickerSave := time.NewTicker(time.Duration(reportInterval) * time.Second)
	fmt.Println("Report interval is", reportInterval)

	for {
		select {
		case <-tickerSave.C:
			logger.Info("Tick...")
			go save.SaveMetrics(m)
		case <-tickerSend.C:
			logger.Info("Tock...")
			go send.SendMetrics()
		}
	}
}
