package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"text/template"

	"go-yandex-metrics/cmd/config"
	"go-yandex-metrics/internal/storage"

	"github.com/go-chi/chi/v5"
)

type HTMLPage struct {
	//	Title string
	HTML string
}

var ErrItemNotFound = errors.New("item not found")
var MetricsDB = storage.NewMemStorage()

func RunServer() error {
	cfg, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	// fmt.Println(cfg)
	r := chi.NewRouter()
	r.Post("/update/{mtype}/{mname}/{mvalue}", UpdateHandler)
	r.Get("/value/{mtype}/{mname}", GetHandler)
	r.Get("/", IndexHandler)

	if err := http.ListenAndServe(cfg.Server.Host, r); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has encountered an error %w", err)
		}
	}
	return nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	cwd, _ := os.Getwd()
	os := runtime.GOOS
	html := HTMLPage{getAllMetrics()}
	// html := HTMLPage{"All Metrics", getAllMetrics()}
	pathToT := filepath.Join(cwd, "cmd","server","templates","metrics.txt")
	// pathToT = filepath.Join(cwd, "cmd/server/templates/metrics.html")
	if os == "windows" {
		pathToT = "/dev/projects/yandex-practicum/go-yandex-metrics/cmd/server/templates/metrics.txt"
		// pathToT = "/dev/projects/yandex-practicum/go-yandex-metrics/cmd/server/templates/metrics.html"
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
	for mName, mValue := range MetricsDB.Gauge {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range MetricsDB.Counter {
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
	// html := HTMLPage{"Metric Data for " + mType + " " + mName, mValue}
	html := HTMLPage{mValue}
	if mValue == "" {
		html = HTMLPage{}
		// html = HTMLPage{"No data", "No data available yet"}
	}

	pathToT := filepath.Join(cwd, "cmd","server","templates","metrics.txt")

	if os == "windows" {
		pathToT = "/dev/projects/yandex-practicum/go-yandex-metrics/cmd/server/templates/metrics.txt"
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
	var html string

	switch mType {
	case "gauge":
		if mValue, ok := MetricsDB.Gauge[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case "counter":
		if mValue, ok := MetricsDB.Counter[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}

	// fmt.Println(mType, mName)
	// html := "Metric type: <b>" + mType + "</b><br>"
	// html += "Metric name: <b>" + mName + "</b><br>"
	// html += "Metric value: <b>" + mValue + "</b><br>"
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	// logger.Info("Got request: " + r.RequestURI)

	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue := chi.URLParam(r, "mvalue")

	// logger.Info("Counter data: " + mType + mName + mValue)

	if mType == "gauge" || mType == "counter" {
		MetricsDB.Save(mType, mName, mValue, w)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}
