package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"go-yandex-metrics/cmd/server/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()
	cfg, err := api.NewServerConfig()
	if err != nil {
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			fmt.Sprintf("failed to create config: %v", err),
		)
	}

	store := storage.NewMemStorage()
	server := api.NewServer(cfg.HTTPServer, store)

	err = server.Start(ctx)
	if err != nil {
		log.Fatal("failed to start server %w", err)
	}
}
