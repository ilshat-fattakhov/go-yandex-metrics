package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"go-yandex-metrics/internal/storage"
)

const (
	ContentType     = "Content-Type"
	applicationJSON = "application/json"
	oSuserRW        = 0o600
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := s.getAllMetrics()

	t := s.tpl
	var doc bytes.Buffer
	err := t.Execute(&doc, allMetrics)
	if err != nil {
		s.logger.Info(fmt.Sprintf("an error occured processing template data: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentType, "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	html := doc.String()

	_, err = w.Write([]byte(html))
	if err != nil {
		s.logger.Info(fmt.Sprintf("an error occured writing to browser: %v", err))
	}
}

func (s *Server) getAllMetrics() string {
	html := "<h3>Gauge:</h3>"
	for mName, mValue := range s.store.MemStore.Gauge {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range s.store.MemStore.Counter {
		html += (mName + ":" + strconv.FormatInt(mValue, 10) + "<br>")
	}
	return html
}

func (s *Server) GetHandler(lg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedJSON := false
		contentType := r.Header.Get("Content-Type")
		if contentType == applicationJSON {
			requestedJSON = true
		}

		if requestedJSON {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				lg.Info("error reading request body", zap.Error(err))
			}

			w.Header().Set(ContentType, applicationJSON)

			var m storage.Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mValue, err := s.getSingleMetric(m.MType, m.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				if mValue == "" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var metric = storage.Metrics{}

				switch m.MType {
				case GaugeType:
					mValue, err := strconv.ParseFloat(mValue, 64)
					if err != nil {
						s.logger.Info(fmt.Sprintf("failed to convert string to float64 value: %v", err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = storage.Metrics{ID: m.ID, MType: m.MType, Value: &mValue}
				case CounterType:
					mValue, err := strconv.ParseInt(mValue, 10, 64)
					if err != nil {
						s.logger.Info(fmt.Sprintf("failed to convert string to int64 value: %v", err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = storage.Metrics{ID: m.ID, MType: m.MType, Delta: &mValue}
				}

				w.WriteHeader(http.StatusOK)

				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					s.logger.Info(fmt.Sprintf("failed to JSON encode gauge metric: %v", err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
		} else {
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
				s.logger.Info(fmt.Sprintf("an error occured processing template data: %v", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set(ContentType, "text/html")
			w.Header().Set("charset", "utf-8")
			w.WriteHeader(http.StatusOK)

			html := doc.String()

			_, err = w.Write([]byte(html))
			if err != nil {
				s.logger.Info(fmt.Sprintf("an error occured writing to browser: %v", err))
			}
		}
	}
}

func (s *Server) getSingleMetric(mType, mName string) (string, error) {
	var html string
	var ErrItemNotFound = errors.New("item not found")

	switch mType {
	case GaugeType:
		if mValue, ok := s.store.MemStore.Gauge[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case CounterType:
		if mValue, ok := s.store.MemStore.Counter[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}
}

func (s *Server) UpdateHandler(lg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/update/" && r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Info(fmt.Sprintf("aerror reading request body: %v", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set(ContentType, applicationJSON)

			var m storage.Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				lg.Info("Got error decoding JSON request")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mType := m.MType
			mName := m.ID
			var mValueFloat string
			var mValueInt string

			switch mType {
			case GaugeType:
				mValueFloat = strconv.FormatFloat(*m.Value, 'f', -1, 64)
				storage.SaveMetric(s.store, mType, mName, mValueFloat, w)

				w.WriteHeader(http.StatusOK)

				metric := storage.Metrics{ID: mName, MType: GaugeType, Value: m.Value}

				metricJSON, err := json.Marshal(metric) // metricJSON is of type []byte
				if err != nil {
					s.logger.Info(fmt.Sprintf("failed to marshal metric: %v", err))
					return
				}
				w.Header().Set("Content-Length", bytes.NewBuffer(metricJSON).String())

				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					s.logger.Info(fmt.Sprintf("failed to JSON encode metric: %v", err))
					return
				}

			case CounterType:
				mValueInt = strconv.FormatInt(*m.Delta, 10)
				storage.SaveMetric(s.store, mType, mName, mValueInt, w)
				w.WriteHeader(http.StatusOK)

				metric := storage.Metrics{ID: mName, MType: CounterType, Delta: m.Delta}

				metricJSON, err := json.Marshal(metric) // metricJSON is of type []byte
				if err != nil {
					s.logger.Info("got error marshalling metrics to JSON")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Length", bytes.NewBuffer(metricJSON).String())

				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					s.logger.Info(fmt.Sprintf("failed to JSON encode metric: %v", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

			default:
				lg.Info("wrong metric type - neither gauge nor counter")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			mType := chi.URLParam(r, "mtype")
			mName := chi.URLParam(r, "mname")
			mValue := chi.URLParam(r, "mvalue")

			if mType == GaugeType || mType == CounterType {
				storage.SaveMetric(s.store, mType, mName, mValue, w)
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}
