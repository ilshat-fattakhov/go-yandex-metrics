package api

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"

	"go-yandex-metrics/internal/config"

	"github.com/go-chi/chi/v5"
)

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := getAllMetrics(s)

	t := s.tpl
	var doc bytes.Buffer
	err := t.Execute(&doc, allMetrics)
	if err != nil {
		log.Printf("an error occured processing template data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	html := doc.String()
	_, err = w.Write([]byte(html))
	if err != nil {
		log.Printf("an error occured writing to browser: %v", err)
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
	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue, err := s.getSingleMetric(mType, mName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	t := s.tpl
	var doc bytes.Buffer
	err = t.Execute(&doc, mValue)
	if err != nil {
		log.Printf("an error occured processing template data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	html := doc.String()
	_, err = w.Write([]byte(html))
	if err != nil {
		log.Printf("an error occured writing to browser: %v", err)
	}
}

func (s *Server) getSingleMetric(mType, mName string) (string, error) {
	var html string
	var ErrItemNotFound = errors.New("item not found")

	switch mType {
	case config.GaugeType:
		if mValue, ok := s.store.Gauge[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case config.CounterType:
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

	if mType == config.GaugeType || mType == config.CounterType {
		s.Save(mType, mName, mValue, w)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *Server) Save(mType, mName, mValue string, w http.ResponseWriter) {
	switch mType {
	case config.CounterType:
		s.saveCounter(mName, mValue, w)
	case config.GaugeType:
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
