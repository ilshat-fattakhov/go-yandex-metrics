package api

import (
	"bytes"
	"encoding/json"
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

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string   `json:"id"`              // имя метрики
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := s.store.GetAllMetrics()

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
				s.logger.Info("error reading request body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set(ContentType, applicationJSON)

			var m Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				s.logger.Info(fmt.Sprintf("failed to unmarshal body: %v", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mValue, err := s.store.GetMetric(m.MType, m.ID)
			if err != nil {
				s.logger.Info(fmt.Sprintf("failed to get metric: %v", err))
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				if mValue == "" {
					s.logger.Info(fmt.Sprintf("metric value is empty: %v", err))
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var metric = Metrics{}

				switch m.MType {
				case GaugeType:
					mValue, err := strconv.ParseFloat(mValue, 64)
					if err != nil {
						s.logger.Info(fmt.Sprintf("failed to convert string to float64 value: %v", err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = Metrics{ID: m.ID, MType: m.MType, Value: &mValue}
				case CounterType:
					mValue, err := strconv.ParseInt(mValue, 10, 64)
					if err != nil {
						s.logger.Info(fmt.Sprintf("failed to convert string to int64 value: %v", err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = Metrics{ID: m.ID, MType: m.MType, Delta: &mValue}
				}

				w.WriteHeader(http.StatusOK)

				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					s.logger.Info(fmt.Sprintf("failed to JSON encode gauge metric: %v", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}
		} else {
			mType := chi.URLParam(r, "mtype")
			mName := chi.URLParam(r, "mname")
			mValue, err := s.store.GetMetric(mType, mName)
			if err != nil {
				s.logger.Info(fmt.Sprintf("metric not found: %v", err))
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

			html := doc.String()

			_, err = w.Write([]byte(html))
			if err != nil {
				s.logger.Info(fmt.Sprintf("an error occured writing to browser: %v", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set(ContentType, "text/html")
			w.Header().Set("charset", "utf-8")
			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

func (s *Server) UpdateHandler(lg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/update/" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Info(fmt.Sprintf("error reading request body: %v", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set(ContentType, applicationJSON)

			var m Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				s.logger.Info(fmt.Sprintf("error decoding JSON request: %v", err))
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
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueFloat); err != nil {
					s.logger.Info(fmt.Sprintf("error saving metric: %v", mType))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)

				metric := Metrics{ID: mName, MType: GaugeType, Value: m.Value}

				metricJSON, err := json.Marshal(metric) // metricJSON is of type []byte
				if err != nil {
					s.logger.Info(fmt.Sprintf("failed to marshal metric: %v", err))
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

			case CounterType:
				mValueInt = strconv.FormatInt(*m.Delta, 10)
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueInt); err != nil {
					s.logger.Info(fmt.Sprintf("error saving metric: %v", mType))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)

				metric := Metrics{ID: mName, MType: CounterType, Delta: m.Delta}

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
			mType := chi.URLParam(r, "mtype")
			mName := chi.URLParam(r, "mname")
			mValue := chi.URLParam(r, "mvalue")

			if mType == GaugeType || mType == CounterType {
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValue); err != nil {
					s.logger.Info(fmt.Sprintf("error saving metric: %v", mType))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}
