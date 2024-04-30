package main

import (
	"fmt"
	"go-yandex-metrics/cmd/agent/storage"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// данная структура хранит метрики
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var m = MemStorage{
	counter: make(map[string]int64),
	gauge:   make(map[string]float64),
}

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func updateRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/update/{mtype}/{mname}/{mvalue}", updateHandler) // POST /update/counter/PollCount/1
	r.Get("/value/{mtype}/{mname}", getHandler)               // GET /value/counter/PollCount
	r.Get("/", indexHandler)
	return r
}
func main() {
	log.Fatal(http.ListenAndServe(":8080", updateRouter()))
}

func isset(arr []string, index int) bool {
	return (len(arr) > index)
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
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logger.Info(" with mType: " + mType)

	if mType == "gauge" || mType == "counter" {
		mValue := getCounterValue(mType, mName)
		if mValue == "" {
			// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.Info("Sending metrics to browser. Type: " + mType + " Name: " + mName + " Value: " + mValue)

			w.Header().Set("content-type", "text/plain")
			w.Header().Set("charset", "utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mValue + "\n"))

		}
		return

	} else {
		// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.
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

	if mType == "" && mName == "" && mValue == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	logger.Info(" with mType: " + mType)

	if mType == "gauge" || mType == "counter" {
		if mName != "" {
			if mValue != "" {
				err := m.save(mType, mName, mValue)
				if err != nil {
					// При попытке передать запрос с некорректным значением возвращать http.StatusBadRequest
					w.WriteHeader(http.StatusBadRequest)
					return
				} else {
					logger.Info("Saving metrics. Type: " + mType + " Name: " + mName + " Value: " + mValue)
					w.Header().Set("content-type", "text/plain")
					w.Header().Set("charset", "utf-8")
					w.WriteHeader(http.StatusOK)
					return
				}
			} else {
				// При попытке передать запрос с пустым значением возвращать http.StatusBadRequest
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			// При попытке передать запрос без имени метрики возвращать http.StatusNotFound
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		logger.Info(" and we are returning StatusBadRequest ")

		// При попытке передать запрос с некорректным типом метрики возвращать http.StatusBadRequest
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (m *MemStorage) save(t, n, v string) error {

	if t == "counter" {

		// в случае если мы по какой-то причине получили число с плавающей точкой
		vFloat64, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		vInt64 := int64(vFloat64)
		// новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу
		m.counter[n] += vInt64
		return nil

	} else if t == "gauge" {

		vFloat64, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		// новое значение должно замещать предыдущее
		m.gauge[n] = vFloat64
		return nil

	}
	return nil
}
