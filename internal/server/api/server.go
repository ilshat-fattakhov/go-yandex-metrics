package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/storage"
)

type ServerCfg struct {
	Host string
}

type Server struct {
	store  storage.MemStorage
	router *chi.Mux
	cfg    config.ServerCfg
}

func NewServer(cfg config.ServerCfg, store *storage.MemStorage) *Server {
	srv := &Server{
		cfg:    cfg,
		store:  *store,
		router: chi.NewRouter(),
	}

	srv.routes()

	return srv
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
	s.router.Route("/", func(r chi.Router) {
		r.Get("/", s.IndexHandler)
		r.Get("/value/{mtype}/{mname}", s.GetHandler)
		r.Post("/update/{mtype}/{mname}/{mvalue}", s.UpdateHandler)
	})
}
