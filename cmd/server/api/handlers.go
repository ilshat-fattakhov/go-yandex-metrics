package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
)

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	allMetrics := getAllMetrics(s)
	tpl := `{{.}}`
	t, err := template.New("All Metrics").Parse(tpl)
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("an error occured parsing template: %v", err),
		)
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, allMetrics)
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("an error occured converting template data: %v", err),
		)
	}

	html := doc.String()
	_, err = w.Write([]byte(html))
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("an error occured writing to browser: %v", err),
		)
	}
}

func getAllMetrics(s *Server) string {
	html := "<h3>Gauge:</h3>"
	for mName, mValue := range s.store.Gauge {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range s.store.Counter {
		html += (mName + ":" + strconv.FormatInt(mValue, 10) + "<br>")
	}

	return html
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// var ErrItemNotFound = errors.New("item not found")

	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue, err := s.getSingleMetric(mType, mName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
		// if errors.Is(err, ErrItemNotFound) { // по како-то непонятной причине этот код (сравнение ошибки с ErrItemNotFound) не срабоатывает никак
		//	http.Error(w, err.Error(), http.StatusNotFound)
		//	return
		// }
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	tpl := `{{.}}`
	t, err := template.New("Single Metric").Parse(tpl)
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("an error occured parsing template: %v", err),
		)
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, mValue)
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("an error occured converting template data: %v", err),
		)
	}
	html := doc.String()
	_, err = w.Write([]byte(html))
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("an error occured writing to browser: %v", err),
		)
	}
}

func (s *Server) getSingleMetric(mType, mName string) (string, error) {
	var html string
	var ErrItemNotFound = errors.New("item not found")

	switch mType {
	case "gauge":
		if mValue, ok := s.store.Gauge[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case "counter":
		if mValue, ok := s.store.Counter[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}
}

func (s *Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue := chi.URLParam(r, "mvalue")

	if mType == "gauge" || mType == "counter" {
		s.Save(mType, mName, mValue, w)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}

func (s *Server) Save(mType, mName, mValue string, w http.ResponseWriter) {
	switch mType {
	case "counter":
		s.saveCounter(mName, mValue, w)
	case "gauge":
		s.saveGauge(mName, mValue, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Server) saveCounter(mName, mValue string, w http.ResponseWriter) {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	s.store.Counter[mName] += vInt64
}

func (s *Server) saveGauge(mName, mValue string, w http.ResponseWriter) {
	vFloat64, err := strconv.ParseFloat(mValue, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.store.Gauge[mName] = vFloat64
}
