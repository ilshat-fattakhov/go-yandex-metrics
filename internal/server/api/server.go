package api

import (
	"errors"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/storage"
)

type ServerCfg struct {
	Host string
}

type Server struct {
	store   *storage.MemStorage
	router  *chi.Mux
	cfg     config.ServerCfg
	tSingle *template.Template
	tAll    *template.Template
}

func NewServer(cfg config.ServerCfg, store *storage.MemStorage) *Server {
	srv := &Server{
		store:   store,
		router:  chi.NewRouter(),
		cfg:     cfg,
		tSingle: createTemplate("one"),
		tAll:    createTemplate("all"),
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

func createTemplate(w string) *template.Template {
	var tplt *template.Template
	switch w {
	case "one":
		tpl := `{{.}}`
		t, err := template.New("Single Metric").Parse(tpl)
		if err != nil {
			log.Fatalf("an error occured parsing template for single metric: %v", err)
		}
		tplt = t
	case "all":
		tpl := `{{.}}`
		t, err := template.New("All Metrics").Parse(tpl)
		if err != nil {
			log.Fatalf("an error occured parsing template for all metrics: %v", err)

		}
		tplt = t
	}
	return tplt
}
