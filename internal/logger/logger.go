package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(logFile string) *zap.Logger {
	logToFile := false

	if logToFile {
		writerSyncer := getLogWriter(logFile)
		encoder := getEncoder()
		core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)
		logger := zap.New(core)
		return logger
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Can't initialize zap logger: %v", err)
	}
	// defer func() {
	//	if err := logger.Sync(); err != nil {
	//		log.Printf("failed to sync logger: %v", err)
	//		return
	//	}
	// }()
	return logger
}

type EncoderConfig struct {
	// Set the keys used for each log entry. If any key is empty, that portion
	// of the entry is omitted.
	CallerKey     string `json:"callerKey" yaml:"callerKey"`
	FunctionKey   string `json:"functionKey" yaml:"functionKey"` // this needs to be set
	StacktraceKey string `json:"stacktraceKey" yaml:"stacktraceKey"`
}

func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}

func getLogWriter(logFile string) zapcore.WriteSyncer {
	file, _ := os.Create(logFile)
	return zapcore.AddSync(file)
}

func Logger(lg *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				lg.Info("Error getting body")
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			t1 := time.Now()
			defer func() {
				lg.Info("Request info",
					zap.String("URI", r.RequestURI),
					zap.String("method", r.Method),
					zap.String("body", string(body)),
					zap.Duration("time", time.Since(t1)),
				)
			}()
			defer func() {
				lg.Info("Response info",
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
