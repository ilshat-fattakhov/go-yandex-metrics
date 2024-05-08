package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"text/template"

	"go-yandex-metrics/internal/storage"

	"github.com/go-chi/chi/v5"
)

type HTMLPage struct {
	Title string
	HTML  string
}

var ErrItemNotFound = errors.New("item not found")

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	cwd, _ := os.Getwd()
	os := runtime.GOOS
	html := HTMLPage{"All Metrics", getAllMetrics()}
	pathToT := ""
	if os == "windows" {
		pathToT = "/dev/projects/yandex-practicum/go-yandex-metrics/cmd/server/templates/metrics.html"
	} else {
		pathToT = filepath.Join(cwd, "./cmd/server/template/metrics.html")
	}
	t, err := template.ParseFiles(pathToT)
	if err != nil {
		log.Printf("error parsing template file: %v", err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)
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
	mValue, err := getSingleMetric(mType, mName)

	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}
	cwd, _ := os.Getwd()
	os := runtime.GOOS
	html := HTMLPage{"Metric Data for " + mType + " " + mName, mValue}
	if mValue == "" {
		html = HTMLPage{"No data", "No data available yet"}
	}
	pathToT := ""
	if os == "windows" {
		pathToT = "/dev/projects/yandex-practicum/go-yandex-metrics/cmd/server/templates/metrics.html"
	} else {
		pathToT = filepath.Join(cwd, "./cmd/server/template/metrics.html")
	}

	t, err := template.ParseFiles(pathToT)
	if err != nil {
		log.Printf("error parsing template file for individual metrics: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)
	err = t.Execute(w, html)
	if err != nil {
		log.Printf("error creating template for individual metrics: %v", err)
	}
}

func getSingleMetric(mType, mName string) (string, error) {
	mValue := storage.Mem.Get(mType, mName)

	html := "Metric type: <b>" + mType + "</b><br>"
	html += "Metric name: <b>" + mName + "</b><br>"
	html += "Metric value: <b>" + mValue + "</b><br>"

	if mValue == "" {
		return "", ErrItemNotFound
	} else {
		return html, nil
	}
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
