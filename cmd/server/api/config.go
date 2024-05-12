package api

import (
	"flag"
	"os"

	"github.com/go-chi/chi/v5"

	"go-yandex-metrics/internal/storage"
)

type HTTPServer struct {
	Host string
}

type Configuration struct {
	HTTPServer
}

type Server struct {
	cfg    HTTPServer
	store  storage.MemStorage
	router *chi.Mux
}

func NewServerConfig() (Configuration, error) {
	var cfg Configuration

	defaultRunAddr := "localhost:8080"
	var flagRunAddr string

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}
	return cfg, nil
}
