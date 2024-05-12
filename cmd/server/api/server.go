package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/internal/storage"
)

func NewServer(cfg HTTPServer, store *storage.MemStorage) *Server {
	srv := &Server{
		cfg:    cfg,
		store:  *store,
		router: chi.NewRouter(),
	}

	srv.routes()

	return srv
}
func (s *Server) Start(ctx context.Context) error {
	server := http.Server{
		Addr:    s.cfg.Host,
		Handler: s.router,
	}
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has encountered an error %w", err)
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
