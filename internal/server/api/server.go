package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/gzip"
	"go-yandex-metrics/internal/logger"
	"go-yandex-metrics/internal/storage"
)

type ServerCfg struct {
	Host            string `json:"host"`
	FileStoragePath string `json:"file_storage_path"`
	StoreInterval   uint64 `json:"store_interval"`
	Restore         bool   `json:"restore"`
}

type Server struct {
	logger *zap.Logger
	router *chi.Mux
	tpl    *template.Template
	store  *storage.MemStorage
	cfg    config.ServerCfg
}

func NewServer(cfg config.ServerCfg, store *storage.MemStorage) (*Server, error) {
	tpl, err := createTemplate()
	lg := logger.InitLogger("server.log")

	if err != nil {
		return nil, fmt.Errorf("an error occured parsing metrics template: %w", err)
	}

	srv := &Server{
		store:  store,
		router: chi.NewRouter(),
		cfg:    cfg,
		tpl:    tpl,
		logger: lg,
	}
	srv.routes()

	if cfg.Restore {
		err := LoadMetrics(srv)
		if err != nil {
			lg.Info("Failed to load metrics")
		}
	}

	return srv, nil
}

func (s *Server) Start() error {
	s.logger.Info("Storage path: " + s.cfg.FileStoragePath)
	s.logger.Info("Restore on start: " + strconv.FormatBool(s.cfg.Restore))
	s.logger.Info("Store interval: " + strconv.FormatUint(s.cfg.StoreInterval, 10))

	server := http.Server{
		Addr:    s.cfg.Host,
		Handler: s.router,
	}

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server has encountered an error %v", err)
			return nil
		}
	}

	return nil
}

func (s *Server) routes() {
	lg := s.logger

	s.router.Route("/", func(r chi.Router) {
		r.Use(logger.Logger(lg))
		// r.Use(gzip.GzipHandle())
		// r.Use(gzip.GzipMiddleware())
		// r.Get("/", s.IndexHandler)

		r.Get("/value/{mtype}/{mname}", s.GetHandler(lg))
		r.Post("/update/{mtype}/{mname}/{mvalue}", s.UpdateHandler(lg))

		r.Get("/", gzip.GzipMiddleware(s.IndexHandler))
		r.Post("/value/", gzip.GzipMiddleware(s.GetHandlerJSON(lg)))
		r.Post("/update/", gzip.GzipMiddleware(s.UpdateHandlerJSON(lg)))
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

func LoadMetrics(s *Server) error {
	data, err := os.ReadFile(s.cfg.FileStoragePath)
	if err != nil {
		s.logger.Info("Cannot read storage file")
	}
	if err := json.Unmarshal(data, s.store); err != nil {
		s.logger.Info("Cannot unmarshal storage file")
	}
	return nil
}
