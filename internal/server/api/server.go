package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/gzip"
	logger "go-yandex-metrics/internal/server/middleware"
	"go-yandex-metrics/internal/storage"
)

type Server struct {
	router *chi.Mux
	tpl    *template.Template
	logger *zap.Logger
	store  storage.Storage
	cfg    *config.ServerCfg
}

type ServerCfg struct {
	Host       string `json:"host"`
	HashKey    string
	StorageCfg StorageCfg
}

type StorageCfg struct {
	FileStoragePath string
	StoreInterval   uint64 `json:"store_interval"`
	Restore         bool   `json:"restore"`
}

func NewServer(cfg *config.ServerCfg, store storage.Storage) (*Server, error) {
	lg, err := logger.InitLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	tpl, err := createTemplate()
	if err != nil {
		lg.Info("got error parsing metrics template")
		return nil, fmt.Errorf("an error occured parsing metrics template: %w", err)
	}

	srv := &Server{
		router: chi.NewRouter(),
		tpl:    tpl,
		logger: lg,
		store:  store,
		cfg:    cfg,
	}

	srv.routes()

	return srv, nil
}

func (s *Server) Start(cfg *config.ServerCfg) error {
	server := http.Server{
		Addr:    cfg.Host,
		Handler: s.router,
	}

	saveData(s)

	s.logger.Info("starting server")
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has encountered an errors: %w", err)
		}
	}

	return nil
}

func (s *Server) routes() {
	lg := s.logger

	s.router.Route("/", func(r chi.Router) {
		r.Use(logger.Logger(lg))
		r.Use(s.GzipMiddleware())

		r.Get("/", s.IndexHandler)
		r.Get("/ping", s.PingHandler)

		r.Get("/value/{mtype}/{mname}", s.GetHandler(lg))
		r.Post("/value/", s.GetHandler(lg))

		r.Post("/update/{mtype}/{mname}/{mvalue}", s.UpdateHandler(lg))
		r.Post("/updates/", s.UpdatesHandler(lg))
		r.Post("/update/", s.UpdateHandler(lg))
	})
}

func createTemplate() (*template.Template, error) {
	tpl := `{{.}}`
	t, err := template.New("Metrics Template").Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("an error occured parsing metrics template: %w", err)
	}
	return t, nil
}

func saveData(s *Server) {
	if s.cfg.StorageCfg.StoreInterval != 0 && s.cfg.StorageCfg.FileStoragePath != "" {
		go func() {
			tickerStore := time.NewTicker(time.Duration(s.cfg.StorageCfg.StoreInterval) * time.Second)
			for range tickerStore.C {
				err := storage.SaveMetrics(s.store, s.cfg.StorageCfg.FileStoragePath)
				if err != nil {
					log.Fatal("failed to save metrics: %w", err)
				}
			}
		}()
	}
}

func (s *Server) GzipMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := gzip.NewCompressWriter(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						s.logger.Info("failed to close newCompressWriter: %w", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := gzip.NewCompressReader(r.Body)
				if err != nil {
					s.logger.Info("failed to create newCompressReader: %w", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func() {
					if err := cr.Close(); err != nil {
						s.logger.Info("failed to close newCompressReader: %w", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}()
			}
			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}
