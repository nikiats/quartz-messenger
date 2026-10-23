package main

import (
	"log/slog"
	"os"
	loader "quartz/internal/common/config"
	"quartz/internal/gateway/config"
	"time"
)

const shutdownTimeout = 10 * time.Second

func main() {
	var cfg config.Config
	if err := loader.Load(&cfg, "deploy/gateway.env"); err != nil {
		slog.Error("config load error", "error", err)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}
