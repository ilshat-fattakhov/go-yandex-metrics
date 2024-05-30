package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	logger "go-yandex-metrics/cmd/server/middleware"
	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/gzip"
	"go-yandex-metrics/internal/storage"
)

type ServerCfg struct {
	Host          string `json:"host"`
	StoreInterval uint64 `json:"store_interval"`
	Restore       bool   `json:"restore"`
}

type Server struct {
	router     *chi.Mux
	tpl        *template.Template
	logger     *zap.Logger
	store      *storage.FileStorage
	storageCfg config.StorageCfg
	cfg        config.ServerCfg
}

func NewServer(cfg config.ServerCfg, storageCfg config.StorageCfg, store *storage.FileStorage) (*Server, error) {
	tpl, err := createTemplate()
	lg := logger.InitLogger()

	if err != nil {
		return nil, fmt.Errorf("an error occured parsing metrics template: %w", err)
	}

	srv := &Server{
		store:      store,
		router:     chi.NewRouter(),
		cfg:        cfg,
		tpl:        tpl,
		logger:     lg,
		storageCfg: storageCfg,
	}

	if cfg.Restore {
		store, err = store.Load(storageCfg.FileStoragePath)

		if err != nil {
			lg.Fatal("Failed to load metrics from file", zap.Error(err))
		}
		srv.store = store
	}
	srv.routes()

	return srv, nil
}

func (s *Server) Start() error {
	s.logger.Info("Storage path: " + s.storageCfg.FileStoragePath)
	s.logger.Info("Restore on start: " + strconv.FormatBool(s.cfg.Restore))
	s.logger.Info("Store interval: " + strconv.FormatUint(s.cfg.StoreInterval, 10))

	server := http.Server{
		Addr:    s.cfg.Host,
		Handler: s.router,
	}

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Info(fmt.Sprintf("HTTP server has encountered an error: %v", err))
			return nil
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
