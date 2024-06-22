package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func InitLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"/tmp/metrics.log",
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("can't initialize zap logger: %w", err)
	}

	return logger, nil
}

func Logger(lg *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				lg.Error("error getting request body", zap.Error(err))
				ww.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			t1 := time.Now()
			defer func() {
				lg.Info("Response info",
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			}()
			defer func() {
				lg.Info("Request info",
					zap.String("URI", r.RequestURI),
					zap.String("method", r.Method),
					zap.String("body", string(body)),
					zap.Duration("time", time.Since(t1)),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
