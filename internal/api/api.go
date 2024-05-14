package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/internal/storage"
)

type Agent struct {
	store storage.MemStorage
	cfg   AgentCfg
}

type Server struct {
	store  storage.MemStorage
	router *chi.Mux
	cfg    ServerCfg
}

func NewServer(cfg ServerCfg, store *storage.MemStorage) *Server {
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
			log.Println("HTTP server has encountered an error %w", err)
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

func NewAgent(cfg AgentCfg, store *storage.MemStorage) *Agent {
	agt := &Agent{
		cfg:   cfg,
		store: *store,
	}

	return agt
}
func (a *Agent) Start() error {
	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			err := a.SaveMetrics()
			if err != nil {
				log.Println("failed to save metrics: %w", err)
				return nil
			}
		case <-tickerSend.C:
			err := a.SendMetrics()
			if err != nil {
				log.Println("failed to send metrics: %w", err)
				return nil
			}
		}
	}
}
