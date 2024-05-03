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
var mem = storage.NewMemStorage()

func main() {
	r := chi.NewRouter()
	r.Post("/update/{mtype}/{mname}/{mvalue}", updateHandler) // POST /update/counter/PollCount/1
	r.Get("/value/{mtype}/{mname}", getHandler)               // GET /value/counter/PollCount
	r.Get("/", indexHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
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
	logger.Info("Got type:" + chi.URLParam(r, "mtype") + chi.URLParam(r, "mname"))
	//mValue := mem.Get(w, r)
	mValue := mem.Get(mType, mName, w)

	if mValue == "" {
		logger.Info("Oops!")
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
		mem.Save(mType, mName, mValue, w)
	} else {
		logger.Info("Got URL:" + r.RequestURI)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

}
