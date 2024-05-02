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
var m = storage.NewMemStorage()

func main() {
	router := chi.NewRouter()
	router.Post("/update/{mtype}/{mname}/{mvalue}", updateHandler) // POST /update/counter/PollCount/1
	router.Get("/value/{mtype}/{mname}", getHandler)               // GET /value/counter/PollCount
	router.Get("/", indexHandler)

	//logger.Info("Getting values from chi: " + mType + " Name: " + mName + " Value: " + mValue)

	log.Fatal(http.ListenAndServe(":8080", router))
	//log.Fatal(http.ListenAndServe(":8080", updateRouter()))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		// Принимаем запросы только по протоколу HTTP методом GET
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	// Принимать запрос в формате http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>

	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")

	logger.Info("We have a visitor: " + r.RequestURI)

	if mType == "" && mName == "" {
		logger.Info("No metric type and no metric name")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logger.Info(" with mType: " + mType)

	if mType == "gauge" || mType == "counter" {
		mValue := getCounterValue(mType, mName)
		if mValue == "" {
			logger.Info("При попытке запроса метрики без значения сервер должен возвращать http.StatusNotFound ?")
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.Info("Sending metrics to browser. Type: " + mType + " Name: " + mName + " Value: " + mValue)

			w.Header().Set("content-type", "text/plain")
			w.Header().Set("charset", "utf-8")
			w.Header().Set("content-length", "11")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mValue + "\n"))

		}
		return

	} else {
		// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.
		logger.Info("При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.")
		w.WriteHeader(http.StatusNotFound)
		return
	}

}

func getCounterValue(mType, mName string) string {
	if mType == "gauge" {
		if mValue, ok := storage.GaugeMetrics[mName]; ok {
			return fmt.Sprint(mValue)
		}
	} else if mType == "counter" {
		if mValue, ok := storage.GaugeMetrics[mName]; ok {
			return fmt.Sprint(mValue)
		}
	}
	return ""
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

	logger.Info("We have a visitor: " + r.RequestURI)
	logger.Info("Saving metrics. Type: " + mType + " Name: " + mName + " Value: " + mValue)

	//if mType == "" && mName == "" && mValue == "" {
	//	w.WriteHeader(http.StatusOK) // PROBABLY WRONG!!!
	//	return
	//}
	//logger.Info(" with mType: " + mType)

	if mType == "gauge" || mType == "counter" {
		if mName != "" {
			if mValue != "" {
				err := m.Save(mType, mName, mValue)
				if err != nil {
					// При попытке передать запрос с некорректным значением возвращать http.StatusBadRequest
					w.WriteHeader(http.StatusBadRequest)
					return
				} else {
					logger.Info("Saving metrics. Type: " + mType + " Name: " + mName + " Value: " + mValue)
					w.Header().Set("Content-Length", "0")
					w.Header().Set("Content-Type", "text/plain")
					w.Header().Set("Charset", "utf-8")
					w.WriteHeader(http.StatusOK)
					return
				}
			} else {
				// При попытке передать запрос с пустым значением возвращать http.StatusBadRequest
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			logger.Info("При попытке передать запрос без имени метрики возвращать http.StatusNotFound")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		logger.Info("При попытке передать запрос с некорректным типом метрики возвращать http.StatusBadRequest")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
