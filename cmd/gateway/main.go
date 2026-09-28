package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"quaztz/internal/gateway/config"
	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/messenger"
	privatehttp "quaztz/internal/gateway/transport/http"
	"quaztz/internal/gateway/transport/ws"
	loader "quaztz/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg config.Config
	if err := loader.Load(&cfg, "deploy/gateway.env"); err != nil {
		return err
	}

	client, err := messenger.NewClient(cfg.MessengerURL, cfg.InternalToken)
	if err != nil {
		return err
	}

	registry := connection.NewRegistry(cfg.AccessTTL)

	public := http.NewServeMux()
	public.Handle("GET /ws", ws.NewHandler(client, registry, cfg.OriginPatterns))

	private := privatehttp.NewHandler(registry, cfg.InternalToken)

	errs := make(chan error, 2)

	// запуск WebSocket сервера
	wsServer := &http.Server{
		Addr:              cfg.WSAddr,
		Handler:           public,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- wsServer.ListenAndServe()
	}()

	// запуск сервера приватного API
	privateApiServer := &http.Server{
		Addr:              cfg.PrivateAddr,
		Handler:           private,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- privateApiServer.ListenAndServe()
	}()

	slog.Info("gateway started", "ws", cfg.WSAddr, "private", cfg.PrivateAddr)

	return <-errs
}
