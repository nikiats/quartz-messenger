package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"quaztz/internal/messenger/config"
	"quaztz/internal/messenger/gateway"
	"quaztz/internal/messenger/repository/postgres"
	"quaztz/internal/messenger/transport/http/private"
	"quaztz/internal/messenger/transport/http/public"
	"quaztz/internal/messenger/usecase"
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

	gatewayClient, err := gateway.NewClient(cfg.GatewayURL, cfg.InternalToken)
	if err != nil {
		return err
	}

	messengerRepository := postgres.NewMessenger(pool)
	userRepository := postgres.NewUsers(pool)
	sessionRepository := postgres.NewSessions(pool)

	sessions := usecase.NewSessions(
		sessionRepository,
		userRepository,
		gatewayClient,
		cfg.JWTSecret,
		cfg.AccessTTL,
		cfg.RefreshTTL,
	)
	if err := sessions.LoadRevoked(ctx); err != nil {
		return err
	}

	users := usecase.NewUsers(userRepository)
	chats := usecase.NewChats(messengerRepository)
	messages := usecase.NewMessages(messengerRepository, gatewayClient)

	errs := make(chan error, 2)

	publicApiServer := &http.Server{
		Addr:              cfg.PublicAddr,
		Handler:           public.NewHandler(chats, users, messages, sessions, sessions),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- publicApiServer.ListenAndServe()
	}()

	privateApiServer := &http.Server{
		Addr:              cfg.PrivateAddr,
		Handler:           private.NewHandler(cfg.InternalToken, messages, sessions, sessions),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- privateApiServer.ListenAndServe()
	}()

	slog.Info("messenger started", "public", cfg.PublicAddr, "private", cfg.PrivateAddr)

	return <-errs
}
