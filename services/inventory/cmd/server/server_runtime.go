package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rts/inventory/internal/cache"
	"github.com/rts/inventory/internal/consumer"
	"github.com/rts/inventory/internal/publisher"
	"github.com/rts/inventory/internal/service"
)

func initializeLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
}

func connectDatabase(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("connected to database")
	return pool, nil
}

func connectRedis(ctx context.Context, redisURL string) *cache.RedisCache {
	redisCache := cache.NewRedisCache(redisURL)
	if err := redisCache.Ping(ctx); err != nil {
		slog.Warn("redis not available, running without cache", "error", err)
	} else {
		slog.Info("connected to redis")
	}

	return redisCache
}

func connectEventPublisher(uri, exchangeName, queueName string) *publisher.EventPublisher {
	eventPublisher := publisher.NewEventPublisher(uri, exchangeName, queueName)
	if err := eventPublisher.Connect(); err != nil {
		slog.Warn("failed to connect event publisher to RabbitMQ", "error", err)
	} else {
		slog.Info("event publisher connected to RabbitMQ")
	}

	return eventPublisher
}

func startCatalogConsumer(
	ctx context.Context,
	uri string,
	queueName string,
	inventoryService *service.InventoryService,
) *consumer.CatalogConsumer {
	catalogConsumer := consumer.NewCatalogConsumer(uri, queueName, inventoryService)
	if err := catalogConsumer.Start(ctx); err != nil {
		slog.Warn("failed to start catalog consumer", "error", err)
	} else {
		slog.Info("catalog consumer started")
	}

	return catalogConsumer
}

func startReservationSweeper(ctx context.Context, reservationService *service.ReservationService) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("reservation sweeper stopped")
				return
			case <-ticker.C:
				processed, err := reservationService.ExpireBatch(ctx, 100)
				if err != nil {
					slog.Error("reservation sweeper error", "error", err)
				} else if processed > 0 {
					slog.Info("reservation sweeper processed expired reservations", "count", processed)
				}
			}
		}
	}()

	slog.Info("reservation expiry sweeper started")
}

func newHTTPServer(port int, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func handleGracefulShutdown(server *http.Server, consumerCancel context.CancelFunc, sweeperCancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	consumerCancel()
	sweeperCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
}
