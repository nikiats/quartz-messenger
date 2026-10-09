package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"quaztz/internal/gateway/config"
	"quaztz/internal/gateway/connection"
	privatehttp "quaztz/internal/gateway/transport/http"
	"quaztz/internal/gateway/transport/ws"
	loader "quaztz/internal/platform/config"

	"github.com/nats-io/nats.go"
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

	// соединение с NATS pub/sub системой
	nc, err := nats.Connect(
		cfg.NATSUrl,
		nats.Name("gateway"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second*5),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) { log.Warn("nats disconnected", "err", err) }),
		nats.ReconnectHandler(func(_ *nats.Conn) { log.Info("nats reconnected") }),
	)
	if err != nil {
		return err
	}
	defer nc.Drain()

	// in-memory регистр открытых соединений
	registry := connection.NewRegistry(cfg.AccessTTL)

	messengerClient := NewClient()

	errs := make(chan error, 2)

	// запуск WebSocket сервера
	public := http.NewServeMux()
	public.Handle("GET /ws", ws.NewHandler(registry, cfg.OriginPatterns))
	wsServer := &http.Server{
		Addr:              cfg.WSAddress,
		Handler:           public,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- wsServer.ListenAndServe()
	}()

	// запуск сервера приватного API
	private := privatehttp.NewHandler(registry, cfg.InternalToken)
	privateApiServer := &http.Server{
		Addr:              cfg.PrivateAddr,
		Handler:           private,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- privateApiServer.ListenAndServe()
	}()

	slog.Info("gateway started", "ws", cfg.WSAddress, "private", cfg.PrivateAddr)

	return <-errs
}
