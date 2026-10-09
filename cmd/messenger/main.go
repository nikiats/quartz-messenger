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
	"quaztz/internal/messenger/transport/http/public"
	natstransport "quaztz/internal/messenger/transport/nats"
	"quaztz/internal/messenger/usecase"
	loader "quaztz/internal/platform/config"
	"quaztz/internal/platform/token"

	"github.com/alexedwards/argon2id"
	"github.com/nats-io/nats.go"
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

	// соединение с NATS pub/sub системой
	nc, err := nats.Connect(
		cfg.NATSUrl,
		nats.Name("messenger"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(5*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) { slog.Warn("nats disconnected", "error", err) }),
		nats.ReconnectHandler(func(_ *nats.Conn) { slog.Info("nats reconnected") }),
	)
	if err != nil {
		return err
	}
	defer nc.Drain()

	// клиент к gateway service
	gatewayClient := gateway.NewClient(nc)

	// создание сущностей слоя repository
	messengerRepository := postgres.NewMessenger(pool)
	userRepository := postgres.NewUsers(pool)
	sessionRepository := postgres.NewSessions(pool)

	// usecase: управление сессиями
	sessions := usecase.NewSessions(
		sessionRepository,
		userRepository,
		gatewayClient,
		token.NewIssuer(cfg.JWTSecret),
		token.NewVerifier(cfg.JWTSecret),
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

	errs := make(chan error, 1)

	// запуск публичного API
	publicApiServer := &http.Server{
		Addr:              cfg.PublicAddr,
		Handler:           public.NewHandler(chats, users, messages, sessions, sessions),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- publicApiServer.ListenAndServe()
	}()

	// прием запросов шлюзов
	if err := natstransport.NewServer(nc, messages, sessions).Start(); err != nil {
		return err
	}

	slog.Info("messenger started", "public", cfg.PublicAddr)

	return <-errs
}
