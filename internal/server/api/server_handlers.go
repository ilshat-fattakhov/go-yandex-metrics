package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/logger"
	"go-yandex-metrics/internal/storage"
)

const (
	ContentType     = "Content-Type"
	applicationJSON = "application/json"
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

	w.Header().Set(ContentType, "text/html")
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

func (s *Server) GetHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		defer func() {
			logger.Info("Request info",
				zap.String("URI", r.RequestURI),
				zap.String("method", r.Method),
				zap.String("body", string(body)),
				zap.Duration("time", time.Since(t1)),
			)
		}()
		defer func() {
			logger.Info("Response info",
				zap.Int("status", ww.Status()),
				zap.Int("size", ww.BytesWritten()),
			)
		}()

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
		w.Header().Set(ContentType, "text/html")
		w.Header().Set("charset", "utf-8")
		w.WriteHeader(http.StatusOK)
		html := doc.String()
		_, err = w.Write([]byte(html))
		if err != nil {
			log.Printf("an error occured writing to browser: %v", err)
		}
	}
}

func (s *Server) GetHandlerJSON(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		defer func() {
			logger.Info("Request info",
				zap.String("URI", r.RequestURI),
				zap.String("method", r.Method),
				zap.String("body", string(body)),
				zap.Duration("time", time.Since(t1)),
			)
		}()
		defer func() {
			logger.Info("Response info",
				zap.Int("status", ww.Status()),
				zap.Int("size", ww.BytesWritten()),
			)
		}()

		var m storage.Metrics

		err = json.Unmarshal(body, &m)
		if err != nil {
			logger.Info("Got error decoding JSON request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch m.MType {
		case config.GaugeType:
			if mValue, ok := s.store.Gauge[m.ID]; !ok {
				logger.Info("Gauge value doesn't exist for ID: " + m.ID)
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.Header().Add(ContentType, applicationJSON)
				// w.Header().Add("Content-Length", "0")
				w.WriteHeader(http.StatusOK)

				metric := storage.Metrics{ID: m.ID, MType: config.GaugeType, Value: &mValue}
				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					log.Printf("failed to JSON encode gauge metric: %v", err)
					w.WriteHeader(http.StatusGone)
				}

			}
		case config.CounterType:
			if mDelta, ok := s.store.Counter[m.ID]; !ok {
				logger.Info("Counter value doesn't exist for ID: " + m.ID)
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.Header().Add(ContentType, applicationJSON)
				// w.Header().Add("Content-Length", "0")
				w.WriteHeader(http.StatusOK)

				metric := storage.Metrics{ID: m.ID, MType: config.CounterType, Delta: &mDelta}
				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					log.Printf("failed to JSON encode counter metric: %v", err)
					w.WriteHeader(http.StatusGone)
				}
			}
		default:
			logger.Info("Wrong metric type: '" + m.MType + "'")
			w.WriteHeader(http.StatusNotFound)
		}
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
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Server) Save(mType, mName, mValue string, w http.ResponseWriter) {
	logger := logger.InitLogger()

	switch mType {
	case config.CounterType:
		logger.Info("Saving counter. Name: " + mName + " Value: " + mValue)
		s.saveCounter(mName, mValue, w)
	case config.GaugeType:
		logger.Info("Saving gauge. Name: " + mName + " Value: " + mValue)
		s.saveGauge(mName, mValue, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Server) saveCounter(mName, mValue string, w http.ResponseWriter) {
	logger := logger.InitLogger()

	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		logger.Info("Got error pasring float value for counter metric")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	s.store.Counter[mName] += vInt64
	logger.Info("SAVED counter. Name: " + mName + " Value: " + mValue)
}

func (s *Server) saveGauge(mName, mValue string, w http.ResponseWriter) {
	logger := logger.InitLogger()

	vFloat64, err := strconv.ParseFloat(mValue, 64)

	if err != nil {
		logger.Info("Got error pasring float value for gauge metric")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.store.Gauge[mName] = vFloat64
	logger.Info("SAVED gauge. Name: " + mName + " Value: " + mValue)
}

func (s *Server) UpdateHandlerJSON(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		defer func() {
			logger.Info("Request info",
				zap.String("URI", r.RequestURI),
				zap.String("method", r.Method),
				zap.String("body", string(body)),
				zap.Duration("time", time.Since(t1)),
			)
		}()
		defer func() {
			logger.Info("Response info",
				zap.Int("status", ww.Status()),
				zap.Int("size", ww.BytesWritten()),
			)
		}()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var m storage.Metrics

		logger.Info("Decoding JSON request")
		err = json.Unmarshal(body, &m)
		if err != nil {
			logger.Info("Got error decoding JSON request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mType := m.MType
		mName := m.ID
		var mValueFloat string
		var mValueInt string

		switch mType {
		case config.GaugeType:
			mValueFloat = strconv.FormatFloat(*m.Value, 'f', -1, 64)
			s.Save(mType, mName, mValueFloat, w)
			w.Header().Add(ContentType, applicationJSON)
			// w.Header().Add("Content-Length", "0")
			w.WriteHeader(http.StatusOK)

			metric := storage.Metrics{ID: mName, MType: config.GaugeType, Value: m.Value}
			err := json.NewEncoder(w).Encode(metric)
			if err != nil {
				log.Printf("failed to JSON encode metric: %v", err)
				return
			}
			return
		case config.CounterType:
			mValueInt = strconv.FormatInt(*m.Delta, 10)
			s.Save(mType, mName, mValueInt, w)
			w.Header().Add(ContentType, applicationJSON)
			// w.Header().Add("Content-Length", "0")
			w.WriteHeader(http.StatusOK)

			metric := storage.Metrics{ID: mName, MType: config.CounterType, Delta: m.Delta}
			err := json.NewEncoder(w).Encode(metric)
			if err != nil {
				log.Printf("failed to JSON encode metric: %v", err)
				return
			}
			return
		default:
			logger.Info("Wrong metric type - neither gauge nor counter")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
