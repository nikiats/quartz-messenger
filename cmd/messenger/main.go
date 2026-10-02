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

	"github.com/alexedwards/argon2id"
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

	// создание пула соединений с БД
	pool, err := postgres.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	// автоматическая миграция
	if err := postgres.Migrate(ctx, pool); err != nil {
		return err
	}

	// клиент к gateway service
	gatewayClient, err := gateway.NewClient(cfg.GatewayURL, cfg.InternalToken, cfg.MaxGatewayIdleConns)
	if err != nil {
		return err
	}

	// создание сущностей слоя repository
	messengerRepository := postgres.NewMessenger(pool)
	userRepository := postgres.NewUsers(pool)
	sessionRepository := postgres.NewSessions(pool)

	// usecase: управление сессиями
	sessions := usecase.NewSessions(
		sessionRepository,
		userRepository,
		gatewayClient,
		cfg.JWTSecret,
		cfg.AccessTTL,
		cfg.RefreshTTL,
	)
	// загрузка неистекших отозванных сессий в in-memory
	// хранилище для быстрой проверки при каждом запросе
	if err := sessions.LoadRevoked(ctx); err != nil {
		return err
	}

	// usecase: работа с пользователями, чатами, сообщениями
	passwordParams := argon2id.Params(cfg.Argon2)
	users := usecase.NewUsers(userRepository, &passwordParams)
	chats := usecase.NewChats(messengerRepository)
	messages := usecase.NewMessages(messengerRepository, gatewayClient)

	errs := make(chan error, 2)

	// запуск публичного API
	publicApiServer := &http.Server{
		Addr:              cfg.PublicAddr,
		Handler:           public.NewHandler(chats, users, messages, sessions, sessions),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- publicApiServer.ListenAndServe()
	}()

	// запуск внутреннего API
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
