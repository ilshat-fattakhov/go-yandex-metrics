package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// данная структура хранит метрики
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var m = new(MemStorage)

func main() {

	m.counter = make(map[string]int64)
	m.gauge = make(map[string]float64)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/update/", updateHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		// Принимаем метрики только по протоколу HTTP методом POST
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Принимать данные в формате Content-Type: text/plain
	contentType := r.Header.Get("Content-type")
	if contentType != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	urlPart := strings.Split(r.RequestURI, "/")
	mType := urlPart[2]
	mName := urlPart[3]
	mValue := urlPart[4]

	if mType == "gauge" || mType == "counter" {
		if mName != "" {
			if mValue != "" {
				err := m.save(mType, mName, mValue)
				if err != nil {
					//fmt.Fprint(w, err)
					// При попытке передать запрос с некорректным значением возвращать http.StatusBadRequest
					w.WriteHeader(http.StatusBadRequest)
					return
				} else {
					//fmt.Fprint(w, m)
					w.WriteHeader(http.StatusOK)
					w.Header().Set("content-type", "text/plain")
					w.Header().Set("charset", "utf-8")
					return
				}
			} else {
				// При попытке передать запрос с некорректным значением возвращать http.StatusBadRequest
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			// При попытке передать запрос без имени метрики возвращать http.StatusNotFound
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		// При попытке передать запрос с некорректным типом метрики возвращать http.StatusBadRequest
		w.WriteHeader(http.StatusBadRequest)
	}
	return
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Welcome to Yandex Metrics!")
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
