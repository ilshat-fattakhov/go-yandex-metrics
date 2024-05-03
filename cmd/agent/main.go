package main

import (
	send "go-yandex-metrics/cmd/agent/handlers/send"
	"log/slog"

	save "go-yandex-metrics/cmd/agent/handlers/save"

	//"go-yandex-metrics/cmd/agent/storage"

	"os"
	"os/signal"
	"runtime"
	"time"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	c := make(chan os.Signal, 1)
	signal.Notify(c)

	tickerSave := time.NewTicker(save.PollInterval * time.Second)
	tickerSend := time.NewTicker(send.ReportInterval * time.Second)

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
