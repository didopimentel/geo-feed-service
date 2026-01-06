package main

import (
	"context"
	"errors"
	"geo-feed-service/internal/feed"
	"geo-feed-service/internal/http/handlers"
	"geo-feed-service/internal/ingestion"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	// ============================
	// Config
	// ============================

	httpPort := getEnv("HTTP_PORT", "8080")
	postgresURL := mustEnv("POSTGRES_URL")
	redisURL := mustEnv("REDIS_URL")

	// ============================
	// Logger
	// ============================

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting api",
		"port", httpPort,
	)

	// ============================
	// Context & signals
	// ============================

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// ============================
	// Postgres
	// ============================

	pgPool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		logger.Error("failed to create postgres pool", "err", err)
		os.Exit(1)
	}

	if err := pgPool.Ping(ctx); err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}

	defer pgPool.Close()
	logger.Info("connected to postgres")

	// ============================
	// Redis
	// ============================

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("invalid redis url", "err", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(redisOpts)

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}

	defer redisClient.Close()
	logger.Info("connected to redis")

	// ============================
	// Repositories
	// ============================

	ingestionRepository := ingestion.NewRepository(pgPool)
	feedRepository := feed.NewRepository(pgPool)

	// ============================
	// UseCases and Handlers
	// ============================

	var healthUseCases handlers.HealthAPIUseCases
	ingestionService := ingestion.NewService(ingestionRepository)
	feedService := feed.NewService(feedRepository)

	ucs := handlers.UseCases{
		FeedAPIUseCases:      feedService,
		HealthAPIUseCases:    healthUseCases,
		IngestionAPIUseCases: ingestionService,
	}

	r := handlers.NewServer(ucs)

	// ============================
	// HTTP Server
	// ============================

	server := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	// ============================
	// Graceful shutdown
	// ============================

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", "err", err)
	}

	logger.Info("server stopped gracefully")
}

// ============================
// Helpers
// ============================

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env var: " + key)
	}
	return v
}
