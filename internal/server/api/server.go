package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
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
	cfg    config.ServerCfg
}

type ServerCfg struct {
	Host       string `json:"host"`
	StorageCfg StorageCfg
}

type StorageCfg struct {
	FileStoragePath string
	StoreInterval   uint64 `json:"store_interval"`
	Restore         bool   `json:"restore"`
}

func NewServer(cfg config.ServerCfg) (*Server, error) {
	lg := logger.InitLogger()

	store, err := storage.NewStore(cfg)
	if err != nil {
		lg.Info("got error creating storage")
		return nil, fmt.Errorf("got error creating storage: %w", err)
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

func (s *Server) Start(cfg config.ServerCfg) error {
	server := http.Server{
		Addr:    cfg.Host,
		Handler: s.router,
	}
	saveData(s)

	s.logger.Info("starting server")
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Info(fmt.Sprintf("HTTP server has encountered an error: %v", err))
			return fmt.Errorf("HTTP server has encountered an errors: %w", err)
		}
	}

	return nil
}

func (s *Server) routes() {
	lg := s.logger

	s.router.Route("/", func(r chi.Router) {
		r.Use(logger.Logger(lg))
		r.Use(gzip.GzipMiddleware())

		r.Get("/", s.IndexHandler)
		r.Get("/value/{mtype}/{mname}", s.GetHandler(lg))
		r.Post("/value/", s.GetHandler(lg))

		r.Post("/update/{mtype}/{mname}/{mvalue}", s.UpdateHandler(lg))
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
					s.logger.Info(fmt.Sprintf("failed to store metrics: %v", err))
					log.Panicf("failed to store metrics: %v", err)
				}
			}
		}()
	}
}
