package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string  `json:"id"`              // имя метрики
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (s *Server) PingHandler(w http.ResponseWriter, r *http.Request) {
	contentEncoding := r.Header.Get(acceptEncoding)
	acceptsGzip := strings.Contains(contentEncoding, gzipStr)

	_, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", s.cfg.StorageCfg.DatabaseDSN)
	if err != nil {
		s.logger.Info("could not connect to database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := db.Ping(); err != nil {
		s.logger.Info("unable to reach database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if acceptsGzip {
		w.Header().Set(contentEncoding, gzipStr)
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
		s.logger.Info("an error occured getting a list of metrics: %w", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := s.tpl
	var doc bytes.Buffer
	err = t.Execute(&doc, allMetrics)
	if err != nil {
		s.logger.Info("an error occured processing template data: %w", zap.Error(err))
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
		s.logger.Info("an error occured writing to browser: %w", zap.Error(err))
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

		if requestedJSON {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Info("error reading request body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var m Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				s.logger.Info("failed to unmarshal body: %w", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mValue, err := s.store.GetMetric(m.MType, m.ID)
			if err != nil {
				s.logger.Info("failed to get metric: %w", zap.Error(err))
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				if mValue == "" {
					s.logger.Info("metric value is empty: %w", zap.Error(err))
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var metric = Metrics{}

				switch m.MType {
				case GaugeType:
					mValue, err := strconv.ParseFloat(mValue, 64)
					if err != nil {
						s.logger.Info("failed to convert string to float64 value: %w", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = Metrics{ID: m.ID, MType: m.MType, Value: &mValue}
				case CounterType:
					mValue, err := strconv.ParseInt(mValue, 10, 64)
					if err != nil {
						s.logger.Info("failed to convert string to int64 value: %w", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					metric = Metrics{ID: m.ID, MType: m.MType, Delta: &mValue}
				}

				w.Header().Set(contentTypeStr, applicationJSON)
				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}

				err = json.NewEncoder(w).Encode(metric)
				if err != nil {
					s.logger.Info("failed to JSON encode gauge metric: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			mType := chi.URLParam(r, "mtype")
			mName := chi.URLParam(r, "mname")
			mValue, err := s.store.GetMetric(mType, mName)
			if err != nil {
				s.logger.Info("metric not found: %w", zap.Error(err))
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			t := s.tpl
			var doc bytes.Buffer
			err = t.Execute(&doc, mValue)
			if err != nil {
				s.logger.Info("an error occured processing template data: %w", zap.Error(err))
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
				s.logger.Info("an error occured writing to browser: %w", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set(contentTypeStr, "text/html")
			w.Header().Set("charset", "utf-8")
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
				s.logger.Info("error reading request body: %w", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set(contentTypeStr, applicationJSON)

			var m Metrics

			err = json.Unmarshal(body, &m)
			if err != nil {
				s.logger.Info("error decoding JSON request: %w", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			s.verifyHashHeader(r, w, body)

			mType := m.MType
			mName := m.ID
			var mValueFloat string
			var mValueInt string

			switch mType {
			case GaugeType:
				mValueFloat = strconv.FormatFloat(*m.Value, 'f', -1, 64)
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueFloat); err != nil {
					s.logger.Info("error saving gauge metric: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				metric := Metrics{ID: mName, MType: GaugeType, Value: m.Value}

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(metric)
				if err != nil {
					s.logger.Info("failed to JSON encode metric: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}
				_, err = w.Write(buf.Bytes())
				if err != nil {
					s.logger.Info("failed to write to ResponseWriter in UpdateHandler: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set(contentLengthStr, strconv.Itoa(buf.Len()))
				w.WriteHeader(http.StatusOK)
				return
			case CounterType:
				mValueInt = strconv.FormatInt(*m.Delta, 10)
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueInt); err != nil {
					s.logger.Info("error saving counter metric: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				metric := Metrics{ID: mName, MType: CounterType, Delta: m.Delta}

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(metric)
				if err != nil {
					s.logger.Info("failed to JSON encode metric: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if acceptsGzip {
					w.Header().Set(contentEncStr, gzipStr)
				}
				_, err = w.Write(buf.Bytes())
				if err != nil {
					s.logger.Info("failed to write buffer to ResponseWriter: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				s.addHashHeader(w, buf.Bytes())
				w.Header().Set(contentLengthStr, strconv.Itoa(buf.Len()))
				w.WriteHeader(http.StatusOK)
				return

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
					s.logger.Info("error saving metric: %w", zap.Error(err))
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

func (s *Server) UpdatesHandler(lg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get(acceptEncoding)
		acceptsGzip := strings.Contains(contentEncoding, gzipStr)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Info("error reading request body: %w", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set(contentTypeStr, applicationJSON)

		var m []MetricsToSend

		err = json.Unmarshal(body, &m)
		if err != nil {
			s.logger.Info("error decoding JSON request: %w", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		s.verifyHashHeader(r, w, body)

		for _, b := range m {
			mType := b.MType
			mName := b.ID
			var mValueFloat string
			var mValueInt string

			switch mType {
			case GaugeType:
				mValueFloat = strconv.FormatFloat(b.Value, 'f', -1, 64)
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueFloat); err != nil {
					s.logger.Info("error saving gauge metric: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			case CounterType:
				mValueInt = strconv.FormatInt(b.Delta, 10)
				if err := storage.Storage.SaveMetric(s.store, mType, mName, mValueInt); err != nil {
					s.logger.Info("error saving counter metric: %w", zap.Error(err))
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
			s.logger.Info("failed to write buffer to ResponseWriter in UpdatesHandler: %w", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.addHashHeader(w, body)
		w.Header().Set(contentLengthStr, strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) verifyHashHeader(r *http.Request, w http.ResponseWriter, body []byte) {
	var requestHash = ""
	if v, ok := r.Header["Hashsha256"]; ok {
		requestHash = v[0]
	}
	if requestHash != "" {
		buf := bytes.NewBuffer(body)
		severSideHash := s.calcHash(*buf)

		if severSideHash != requestHash {
			// fmt.Println("Hashes are NOT!!! the same")
			// fmt.Println("In :", requestHash)
			// fmt.Println("Out:", severSideHash)
			s.logger.Info("wrong hash signature")
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			fmt.Println("Hashes are the same")
			fmt.Println("In :", requestHash)
			fmt.Println("Out:", severSideHash)
		}
	}
}

func (s *Server) addHashHeader(w http.ResponseWriter, body []byte) {
	if s.cfg.HashKey != "" {
		buf := bytes.NewBuffer(body)
		responseHash := s.calcHash(*buf)
		w.Header().Add("HashSHA256", responseHash)
	}
}

func (s *Server) calcHash(buf bytes.Buffer) string {
	var secretkey = []byte(s.cfg.HashKey)
	hashSHA256 := hmac.New(sha256.New, secretkey)

	hashSHA256.Write(buf.Bytes())
	bs := hashSHA256.Sum(nil)

	return hex.EncodeToString(bs)
}
