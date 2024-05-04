package main

import (
	"fmt"
	"go-yandex-metrics/internal/storage"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {

	parseFlags()

	r := chi.NewRouter()
	r.Post("/update/{mtype}/{mname}/{mvalue}", updateHandler) // POST /update/counter/PollCount/1
	r.Get("/value/{mtype}/{mname}", getHandler)               // GET /value/counter/PollCount
	r.Get("/", indexHandler)
	fmt.Println("Running server on", flagRunAddr)

	log.Fatal(http.ListenAndServe(flagRunAddr, r))
	//log.Fatal(http.ListenAndServe(":8080", updateRouter()))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed) // Принимаем запросы только по протоколу HTTP методом GET
		return
	}
	logger.Info("Listing all counters with values...")
	// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страницу
	// со списком имён и значений всех известных ему на текущий момент метрик.
	w.Write([]byte("Gauge:\n"))
	for mName, mValue := range storage.GaugeMetrics {
		w.Write([]byte(mName + ":" + fmt.Sprint(mValue) + "\n"))
	}
	w.Write([]byte("Counter:\n"))
	for mName, mValue := range storage.CounterMetrics {
		w.Write([]byte(mName + ":" + fmt.Sprint(mValue) + "\n"))
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		// Принимаем запросы только по протоколу HTTP методом GET
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	logger.Info("Value request URL:" + r.RequestURI)
	logger.Info("Getting " + mType + " metrics with name " + mName) //mValue := mem.Get(w, r)
	mValue := storage.Mem.Get(mType, mName, w)

	if mValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		w.Header().Set("content-type", "text/plain")
		w.Header().Set("charset", "utf-8")
		//w.Header().Set("content-length", "11")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mValue))
		return
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		// Принимаем метрики только по протоколу HTTP методом POST
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue := chi.URLParam(r, "mvalue")
	if mType == "gauge" || mType == "counter" {
		logger.Info("Update request URL:" + r.RequestURI)
		logger.Info("Saving " + mType + " metrics with name " + mName + " and value " + mValue)
		storage.Mem.Save(mType, mName, mValue, w)
	} else {
		logger.Info("Returning 501 for URL: " + r.RequestURI)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

}
