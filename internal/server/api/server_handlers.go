package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"go-yandex-metrics/internal/storage"
)

const (
	contentTypeStr   = "Content-Type"
	contentLengthStr = "Content-Length"
	applicationJSON  = "application/json"
	oSuserRW         = 0o600
)

const (
	GaugeType      string = "gauge"
	CounterType    string = "counter"
	gzipStr        string = "gzip"
	contentEncStr  string = "Content-Encoding"
	acceptEncoding string = "Accept-Encoding"
)

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string   `json:"id"`              // имя метрики
}

type MetricsToSend struct {
	MType string  `json:"type"`
	ID    string  `json:"id"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}

func (s *Server) PingHandler(w http.ResponseWriter, r *http.Request) {
	contentEncoding := r.Header.Get(acceptEncoding)
	acceptsGzip := strings.Contains(contentEncoding, gzipStr)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, s.cfg.StorageCfg.DatabaseDSN)
	if err != nil {
		s.logger.Info("failed to create a connection pool", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := pool.Ping(ctx); err != nil {
		s.logger.Info("could not connect to database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if acceptsGzip {
		w.Header().Set(contentEncStr, gzipStr)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	requestedJSON := false
	contentType := r.Header.Get("Content-Type")
	if contentType == applicationJSON {
		requestedJSON = true
	}

	contentEncoding := r.Header.Get(acceptEncoding)
	acceptsGzip := strings.Contains(contentEncoding, gzipStr)

	allMetrics, err := s.store.GetAllMetrics()
	if err != nil {
		s.logger.Info("an error occured getting a list of metrics:", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := s.tpl
	var doc bytes.Buffer
	err = t.Execute(&doc, allMetrics)
	if err != nil {
		s.logger.Info("an error occured processing template data:", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if requestedJSON {
		w.Header().Set(contentTypeStr, applicationJSON)
	} else {
		w.Header().Set(contentTypeStr, "text/html")
	}
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	if acceptsGzip {
		w.Header().Set(contentEncStr, gzipStr)
	}

	html := doc.String()

	_, err = w.Write([]byte(html))
	if err != nil {
		s.logger.Info("an error occured writing to browser:", zap.Error(err))
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

		contentEncoding := r.Header.Get(acceptEncoding)
		acceptsGzip := strings.Contains(contentEncoding, gzipStr)

		if r.RequestURI == "/value/" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Info("error reading request body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var m Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				s.logger.Info("failed to unmarshal body:", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mValue, err := s.store.GetMetric(m.MType, m.ID)
			if err != nil {
				s.logger.Info("failed to get metric:", zap.Error(err))
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				if mValue == "" {
					s.logger.Info("metric value is empty:", zap.Error(err))
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var metric = Metrics{}

				switch m.MType {
				case GaugeType:
					mValue, err := strconv.ParseFloat(mValue, 64)
					if err != nil {
						s.logger.Info("failed to convert string to float64 value:", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = Metrics{ID: m.ID, MType: m.MType, Value: &mValue}
				case CounterType:
					mValue, err := strconv.ParseInt(mValue, 10, 64)
					if err != nil {
						s.logger.Info("failed to convert string to int64 value:", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = Metrics{ID: m.ID, MType: m.MType, Delta: &mValue}
				}

				if requestedJSON {
					w.Header().Set(contentTypeStr, applicationJSON)
				} else {
					w.Header().Set(contentTypeStr, "text/html")
				}
				w.Header().Set("charset", "utf-8")

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(metric)
				if err != nil {
					s.logger.Info("failed to JSON encode metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}

				_, err = w.Write(buf.Bytes())
				if err != nil {
					s.logger.Info("failed to write to ResponseWriter:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set(contentLengthStr, strconv.Itoa(buf.Len()))
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			mType := chi.URLParam(r, "mtype")
			mName := chi.URLParam(r, "mname")
			mValue, err := s.store.GetMetric(mType, mName)
			if err != nil {
				s.logger.Info("metric not found:", zap.Error(err))
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			t := s.tpl
			var doc bytes.Buffer
			err = t.Execute(&doc, mValue)
			if err != nil {
				s.logger.Info("an error occured processing template data:", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set(contentTypeStr, "text/html")
			w.Header().Set("charset", "utf-8")

			html := doc.String()

			if acceptsGzip {
				w.Header().Set(contentEncStr, gzipStr)
			}

			_, err = w.Write([]byte(html))
			if err != nil {
				s.logger.Info("an error occured writing to browser:", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

func (s *Server) UpdateHandler(lg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get(acceptEncoding)
		acceptsGzip := strings.Contains(contentEncoding, gzipStr)

		if r.RequestURI == "/update/" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Info("error reading request body:", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set(contentTypeStr, applicationJSON)

			var m Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				s.logger.Info("error decoding JSON request:", zap.Error(err))
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
					s.logger.Info("error saving gauge metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				metric := Metrics{ID: mName, MType: GaugeType, Value: m.Value}

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(metric)
				if err != nil {
					s.logger.Info("failed to JSON encode metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}
				_, err = w.Write(buf.Bytes())
				if err != nil {
					s.logger.Info("failed to write to ResponseWriter in UpdateHandler:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set(contentLengthStr, strconv.Itoa(buf.Len()))
				w.WriteHeader(http.StatusOK)
				return
			case CounterType:
				mValueInt = strconv.FormatInt(*m.Delta, 10)
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueInt); err != nil {
					s.logger.Info("error saving counter metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				metric := Metrics{ID: mName, MType: CounterType, Delta: m.Delta}

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(metric)
				if err != nil {
					s.logger.Info("failed to JSON encode metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}
				_, err = w.Write(buf.Bytes())
				if err != nil {
					s.logger.Info("failed to write buffer to ResponseWriter:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set(contentLengthStr, strconv.Itoa(buf.Len()))
				w.WriteHeader(http.StatusOK)
				return

			default:
				lg.Info("wrong metric type - neither gauge nor counter")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			fmt.Println(r.RequestURI)

			mType := chi.URLParam(r, "mtype")
			mName := chi.URLParam(r, "mname")
			mValue := chi.URLParam(r, "mvalue")

			fmt.Println(mType, mName, mValue)

			if mType == GaugeType || mType == CounterType {
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValue); err != nil {
					s.logger.Info("error saving metric:", zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}

func (s *Server) UpdatesHandler(lg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get(acceptEncoding)
		acceptsGzip := strings.Contains(contentEncoding, gzipStr)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Info("error reading request body:", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set(contentTypeStr, applicationJSON)

		var m []MetricsToSend

		err = json.Unmarshal(body, &m)
		if err != nil {
			s.logger.Info("error decoding JSON request:", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, b := range m {
			var mValueFloat string
			var mValueInt string

			switch b.MType {
			case GaugeType:
				mValueFloat = strconv.FormatFloat(b.Value, 'f', -1, 64)
				if err := storage.Storage.SaveMetric(s.store, b.MType, b.ID, mValueFloat); err != nil {
					s.logger.Info("error saving gauge metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			case CounterType:
				mValueInt = strconv.FormatInt(b.Delta, 10)
				if err := storage.Storage.SaveMetric(s.store, b.MType, b.ID, mValueInt); err != nil {
					s.logger.Info("error saving counter metric:", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			default:
				lg.Info("wrong metric type - neither gauge nor counter")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if acceptsGzip {
			w.Header().Set(contentEncStr, gzipStr)
		}
		_, err = w.Write(body)
		if err != nil {
			s.logger.Info("failed to write buffer to ResponseWriter in UpdatesHandler:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set(contentLengthStr, strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
	}
}
