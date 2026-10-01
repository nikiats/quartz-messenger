package main

import (
	"context"
	"log/slog"
	"os"

	"quaztz/internal/messenger/config"
	"quaztz/internal/messenger/repository/postgres"
	loader "quaztz/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("messenger stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg config.Config
	if err := loader.Load(&cfg, "deploy/messenger.env"); err != nil {
		return err
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		return err
	}

	return nil
}
