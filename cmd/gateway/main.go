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

	"quaztz/internal/gateway/config"
	"quaztz/internal/gateway/connection"
	"quaztz/internal/gateway/service"
	natstransport "quaztz/internal/gateway/transport/nats"
	"quaztz/internal/gateway/transport/ws"
	loader "quaztz/internal/platform/config"
	"quaztz/internal/platform/rediskey"
	"quaztz/internal/platform/token"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// загрузка конфигурации
	var cfg config.Config
	if err := loader.Load(&cfg, "deploy/gateway.env"); err != nil {
		return err
	}

	// ID шлюза
	gatewayID := uuid.New()

	// graceful stop (1)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// соединение с NATS pub/sub системой
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

	// соединение с Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	// обновление записи об активности шлюза
	key := rediskey.Gateway(gatewayID)
	if err := rdb.Set(ctx, key, "1", time.Second*2).Err(); err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := rdb.Set(ctx, key, "1", time.Second*2).Err(); err != nil {
					slog.Warn("gateway heartbeat failed", "error", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	defer rdb.Del(context.Background(), key)

	// in-memory регистр открытых соединений
	registry := connection.NewRegistry()

	// сервисы шлюза
	auth := service.NewAuth(rdb, token.NewVerifier(cfg.JWTSecret))
	messenger := service.NewMessenger(nc)
	presence := service.NewPresence(rdb, gatewayID)

	// прием событий мессенджера
	subscriber := natstransport.NewSubscriber(nc, registry)
	if err := subscriber.Start(gatewayID); err != nil {
		return err
	}

	// запуск WebSocket сервера
	public := http.NewServeMux()
	public.Handle("GET /ws", ws.NewHandler(auth, messenger, presence, registry, cfg.OriginPatterns))
	wsServer := &http.Server{
		Addr:              cfg.WSAddress,
		Handler:           public,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		errs <- wsServer.ListenAndServe()
	}()

	slog.Info("gateway started", "ws", cfg.WSAddress)

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	// graceful stop (2)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := wsServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	slog.Info("gateway stopped gracefully")
	return nil
}
