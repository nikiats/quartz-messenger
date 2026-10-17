package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"quartz/internal/common/auth"
	loader "quartz/internal/common/config"
	"quartz/internal/common/token"
	"quartz/internal/gateway/config"
	"quartz/internal/gateway/messenger"
	"quartz/internal/gateway/presence"
	"quartz/internal/gateway/registry"
	natstransport "quartz/internal/gateway/transport/nats"
	"quartz/internal/gateway/transport/ws"
	"quartz/internal/gateway/usecase"
)

const shutdownTimeout = 10 * time.Second

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

	gatewayID := uuid.New()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	nc, err := nats.Connect(
		cfg.NATSUrl,
		nats.Name("gateway"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(5*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) { slog.Warn("nats disconnected", "error", err) }),
		nats.ReconnectHandler(func(_ *nats.Conn) { slog.Info("nats reconnected") }),
	)
	if err != nil {
		return err
	}
	defer nc.Drain()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	hub := usecase.NewHub(registry.New(), presence.New(rdb, gatewayID))
	authentication := auth.NewAuthentication(rdb, token.NewVerifier(cfg.JWTSecret))
	messengerClient := messenger.NewClient(nc)

	if err := natstransport.NewSubscriber(nc, hub).Start(gatewayID); err != nil {
		return err
	}

	public := http.NewServeMux()
	public.Handle("GET /ws", ws.NewHandler(authentication, messengerClient, hub, cfg.OriginPatterns))
	wsServer := &http.Server{
		Addr:              cfg.WSAddress,
		Handler:           public,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		errs <- wsServer.ListenAndServe()
	}()

	slog.Info("gateway started", "id", gatewayID, "ws", cfg.WSAddress)

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := wsServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	hub.CloseAll(shutdownCtx)

	slog.Info("gateway stopped gracefully")
	return nil
}
