package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/cmd/config"
	"go-yandex-metrics/cmd/server/handlers"
)

func main() {
	if err := runServer(); err != nil {
		log.Fatal(err)
	}
}

func runServer() error {
	cfg, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	r := chi.NewRouter()
	r.Post("/update/{mtype}/{mname}/{mvalue}", handlers.UpdateHandler)
	r.Get("/value/{mtype}/{mname}", handlers.GetHandler)
	r.Get("/", handlers.IndexHandler)

	if err := http.ListenAndServe(cfg.Server.Host, r); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has encountered and error %w", err)
		}
	}
	return nil
}
