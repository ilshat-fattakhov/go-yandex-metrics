package main

import (
	"fmt"
	save "go-yandex-metrics/cmd/agent/handlers/save"
	send "go-yandex-metrics/cmd/agent/handlers/send"

	//"go-yandex-metrics/cmd/agent/storage"

	"os"
	"os/signal"
	"runtime"
	"time"
)

func main() {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	c := make(chan os.Signal, 1)
	signal.Notify(c)

	ticker := time.NewTicker(time.Second)
	stop := make(chan bool)

	go func() {
		defer func() { stop <- true }()
		for {
			select {
			case <-ticker.C:
				//fmt.Println("Тик")
				save.SaveMetrics(m)
			case <-stop:
				fmt.Println("Закрытие горутины")
				return
			}
			time.Sleep(save.PollInterval * time.Second)
		}
	}()

	go func() {
		defer func() { stop <- true }()
		for {
			<-ticker.C
			//fmt.Println("Тooooк")
			send.SendMetrics()
			time.Sleep(send.ReportInterval * time.Second)
		}
	}()
	// Блокировка, пока не будет получен сигнал
	<-c
	ticker.Stop()
	// Остановка горутины
	stop <- true
	// Ожидание до тех пор, пока не выполнится
	<-stop
	fmt.Println("Приложение остановлено")
}
