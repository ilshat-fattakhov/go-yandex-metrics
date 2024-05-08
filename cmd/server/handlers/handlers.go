package handlers

import (
	"log"
	"net/http"
	"strconv"
	"text/template"

	"go-yandex-metrics/internal/storage"

	"github.com/go-chi/chi/v5"
)

type HTMLPage struct {
	Title string
	HTML  string
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	html := HTMLPage{"All Metrics", getAllMetrics()}
	t, err := template.ParseFiles("./templates/metrics.html")
	if err != nil {
		log.Printf("error parsing template file: %v", err)
	}

	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, html)
	if err != nil {
		log.Printf("error creating tempate: %v", err)
	}
}

func getAllMetrics() string {
	html := "<h3>Gauge:</h3>"
	for mName, mValue := range storage.Mem.Gauge {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range storage.Mem.Counter {
		html += (mName + ":" + strconv.FormatInt(mValue, 10) + "<br>")
	}
	return html
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")

	mValue := getSingleMetric(mType, mName, w)
	html := HTMLPage{"Metric Data for " + mType + " " + mName, mValue}
	if mValue == "" {
		html = HTMLPage{"No data", "No data available yet"}
	}
	t, err := template.ParseFiles("./templates/metrics.html")
	if err != nil {
		log.Printf("error parsing template file for individual metrics: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, html)
	if err != nil {
		log.Printf("error creating template for individual metrics: %v", err)
	}
}

func getSingleMetric(mType, mName string, w http.ResponseWriter) string {
	mValue := storage.Mem.Get(mType, mName, w)

	html := "Metric type: <b>" + mType + "</b><br>"
	html += "Metric name: <b>" + mName + "</b><br>"
	html += "Metric value: <b>" + mValue + "</b><br>"

	if mValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return ""
	}
	return html
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue := chi.URLParam(r, "mvalue")
	if mType == "gauge" || mType == "counter" {
		storage.Mem.Save(mType, mName, mValue, w)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}
