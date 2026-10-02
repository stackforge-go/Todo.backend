package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stackforge-go/Todo.backend/internal/app"
	"github.com/stackforge-go/Todo.backend/internal/features/auth"
	"github.com/stackforge-go/Todo.backend/internal/features/users"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/hasher"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger/zap"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/postgres/pgx"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/redis/goredis"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/smtp/mailer"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/token/golangjwt"
	"github.com/stackforge-go/Todo.backend/internal/transport/http"
	"github.com/stackforge-go/Todo.backend/internal/transport/rabbitmq"
)

func main() {
	if err := run(); err != err {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// ------------------------------------------------------------------
	// Конфиг приложения
	// ------------------------------------------------------------------
	cfg := app.NewConfigMust()
	time.Local = cfg.TimeZone

	// ------------------------------------------------------------------
	// Shutdown signal
	// ------------------------------------------------------------------
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	// ------------------------------------------------------------------
	// Logger
	// ------------------------------------------------------------------
	log, err := zap.NewLogger(zap.NewConfigMust())
	if err != nil {
		return fmt.Errorf("init application logger: %w", err)
	}
	defer log.Close()

	// ------------------------------------------------------------------
	// Postgres Pool
	// ------------------------------------------------------------------
	pgPool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		return fmt.Errorf("init postgres connection pool: %w", err)
	}
	defer pgPool.Close()
	log.Debug("init postgres connection pool")

	// ------------------------------------------------------------------
	// Redis Pool
	// ------------------------------------------------------------------
	redisPool, err := goredis.NewPool(ctx, goredis.NewConfigMust())
	if err != nil {
		return fmt.Errorf("init redis connection pool: %w", err)
	}
	defer redisPool.Close()
	log.Debug("init redis connection pool")

	// ------------------------------------------------------------------
	// RabbitMQ Server
	// ------------------------------------------------------------------
	rmqServer, err := rabbitmq.NewServer(
		rabbitmq.NewConfigMust(),
		log,
		rabbitmq.RecoverMiddleware(),
		rabbitmq.LoggingMiddleware(),
	)
	if err != nil {
		return fmt.Errorf("init rabbitmq server: %w", err)
	}

	rmqPublisher, err := rmqServer.Publisher()
	if err != nil {
		return fmt.Errorf("init rabbitmq publisher: %w", err)
	}
	_ = rmqPublisher

	// ------------------------------------------------------------------
	// SMTP Client
	// ------------------------------------------------------------------
	smtpClient, err := mailer.New(mailer.NewConfigMust())
	if err != nil {
		return fmt.Errorf("init smtp client: %w", err)
	}
	_ = smtpClient

	// ------------------------------------------------------------------
	// Dependency Injection
	// ------------------------------------------------------------------

	// Shared

	hasher := hasher.NewBcryptHasher()
	tokenIssuer := golangjwt.NewIssuer(golangjwt.NewConfigMust())

	// Repositories

	usersPgRepo := users.NewPgRepository(pgPool)

	// Cached

	usersCachedRepo := users.NewCachedRepository(redisPool, usersPgRepo)

	// Services

	usersUC := users.NewUsecase(
		usersCachedRepo,
		hasher,
	)
	authUC := auth.NewUsecase(
		usersUC,
		tokenIssuer,
	)

	// HTTP Handlers

	usersHTTPHandler := users.NewHTTPHandler(usersUC)
	authHTTPHandler := auth.NewHTTPHandler(
		authUC,
		auth.WithCookieSecure(cfg.Env == "production"),
		auth.WithCookieDomain(cfg.CookieDomain),
	)

	// ------------------------------------------------------------------
	// RabbitMQ router (consumers)
	// ------------------------------------------------------------------
	rmqRouter := rabbitmq.NewRouter()
	rmqServer.RegisterRouter(rmqRouter)

	// ------------------------------------------------------------------
	// HTTP router
	// ------------------------------------------------------------------
	httpRouterV1 := http.NewRouter(http.APIVersion("v1"))
	httpRouterV1.AddRoutes(usersHTTPHandler.Routes())
	httpRouterV1.AddRoutes(authHTTPHandler.Routes())

	// ------------------------------------------------------------------
	// HTTP server
	// ------------------------------------------------------------------
	httpServer := http.NewServer(
		http.NewConfigMust(),
		log,
		http.RecoverMiddleware(),
		http.LoggingMiddleware(),
	)
	httpServer.RegisterRouter(httpRouterV1)

	// ------------------------------------------------------------------
	// Приложение
	// ------------------------------------------------------------------
	application := app.NewApp(
		cfg,
		log,
		httpServer,
		rmqServer,
	)

	// ------------------------------------------------------------------
	// Start приложения
	// ------------------------------------------------------------------
	if err := application.Start(ctx); err != nil {
		log.Error("application start failed", logger.Error(err))
		return err
	}

	// ------------------------------------------------------------------
	// Ждем shutdown signal
	// ------------------------------------------------------------------
	<-ctx.Done()
	log.Info("shutdown signal received")

	// ------------------------------------------------------------------
	// Shutdown приложения
	// ------------------------------------------------------------------
	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancelShutdown()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Error("application shutdown failed", logger.Error(err))
		return err
	}

	return nil
}
