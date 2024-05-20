package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/logger"
	"go-yandex-metrics/internal/storage"
)

type ServerCfg struct {
	Host string
}

type Server struct {
	tpl    *template.Template
	store  *storage.MemStorage
	router *chi.Mux
	cfg    config.ServerCfg
}

func NewServer(cfg config.ServerCfg, store *storage.MemStorage) (*Server, error) {
	tpl, err := createTemplate()
	if err != nil {
		return nil, fmt.Errorf("an error occured parsing metrics template: %w", err)
	}
	srv := &Server{
		store:  store,
		router: chi.NewRouter(),
		cfg:    cfg,
		tpl:    tpl,
	}
	srv.routes()
	return srv, nil
}

func (s *Server) Start() error {
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
	logger := logger.InitLogger()

	s.router.Route("/", func(r chi.Router) {
		r.Get("/", s.IndexHandler)
		r.Get("/value/{mtype}/{mname}", s.GetHandler(logger))
		r.Post("/value/", s.GetHandlerJSON(logger))
		r.Post("/update/", s.UpdateHandlerJSON(logger))
		r.Post("/update/{mtype}/{mname}/{mvalue}", s.UpdateHandler)
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
