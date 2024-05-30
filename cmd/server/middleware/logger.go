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

func InitLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Can't initialize zap logger: %v", err)
	}
	// defer func() {
	//	if err := logger.Sync(); err != nil {
	//		logger.Info(fmt.Sprintf("failed to sync logger: %v", err))
	//		return
	//	}
	// }()
	return logger
}

func Logger(lg *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				lg.Error("Error getting request body", zap.Error(err))
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
		}
		return http.HandlerFunc(fn)
	}
}
