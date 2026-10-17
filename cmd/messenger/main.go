package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"quartz/internal/common/auth"
	loader "quartz/internal/common/config"
	"quartz/internal/common/snowflake"
	"quartz/internal/common/token"
	"quartz/internal/messenger/config"
	"quartz/internal/messenger/gateway"
	"quartz/internal/messenger/repository/cassandra"
	"quartz/internal/messenger/repository/postgres"
	redisrepository "quartz/internal/messenger/repository/redis"
	public "quartz/internal/messenger/transport/http"
	natstransport "quartz/internal/messenger/transport/nats"
	"quartz/internal/messenger/usecase"

	"github.com/alexedwards/argon2id"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
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

	session, err := cassandra.NewSession(cfg.Cassandra.Hosts, cfg.Cassandra.Keyspace, cfg.Cassandra.DisableHostLookup)
	if err != nil {
		return err
	}
	defer session.Close()

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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	ids, err := snowflake.New(ctx, rdb)
	if err != nil {
		return err
	}
	defer ids.Close()

	gatewayClient := gateway.NewClient(nc, rdb, cfg.AccessTTL)

	messengerRepository := cassandra.NewMessenger(session)
	userRepository := postgres.NewUsers(pool)
	sessionRepository := postgres.NewSessions(pool)
	idempotencyRepository := redisrepository.NewIdempotency(rdb, cfg.IdempotencyTTL)

	verifier := token.NewVerifier(cfg.JWTSecret)

	sessions := usecase.NewSessions(
		sessionRepository,
		userRepository,
		gatewayClient,
		auth.NewAuthentication(rdb, verifier),
		token.NewIssuer(cfg.JWTSecret),
		verifier,
		cfg.AccessTTL,
		cfg.RefreshTTL,
	)
	passwordParams := argon2id.Params(cfg.Argon2)
	users := usecase.NewUsers(userRepository, &passwordParams)
	chats := usecase.NewChats(messengerRepository, userRepository, idempotencyRepository, ids)
	messages := usecase.NewMessages(messengerRepository, gatewayClient, idempotencyRepository, ids)

	errs := make(chan error, 1)

	publicApiServer := &http.Server{
		Addr:              cfg.PublicAddr,
		Handler:           public.NewHandler(chats, users, messages, sessions, sessions),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		errs <- publicApiServer.ListenAndServe()
	}()

	if err := natstransport.NewServer(nc, messages, sessions).Start(); err != nil {
		return err
	}

	slog.Info("messenger started", "public", cfg.PublicAddr)

	select {
	case err := <-errs:
		return err
	case err := <-ids.Err():
		return err
	}
}
