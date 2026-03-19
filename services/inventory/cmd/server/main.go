package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/rts/inventory/internal/config"
)

// @title Inventory Service API
// @version 1.0
// @description Real-time inventory management service for the RTS platform
// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations and exit")
	flag.Parse()

	initializeLogger()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Run migrations
	if err := runMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations completed")

	if *migrateOnly {
		slog.Info("migrate-only mode, exiting")
		return
	}

	ctx := context.Background()

	pool, err := connectDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database setup failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisCache := connectRedis(ctx, cfg.RedisURL)
	defer redisCache.Close()

	eventPublisher := connectEventPublisher(cfg.RabbitMQURI, cfg.RabbitMQExchange, cfg.RabbitMQQueue)
	defer eventPublisher.Close()

	app := initializeApplication(cfg, pool, redisCache, eventPublisher)

	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()
	catalogConsumer := startCatalogConsumer(consumerCtx, cfg.RabbitMQURI, cfg.CatalogQueueName, app.inventoryService)
	defer catalogConsumer.Close()

	sweeperCtx, sweeperCancel := context.WithCancel(ctx)
	defer sweeperCancel()
	startReservationSweeper(sweeperCtx, app.reservationService)

	server := newHTTPServer(cfg.Port, app.handler)

	go handleGracefulShutdown(server, consumerCancel, sweeperCancel)

	slog.Info("starting inventory service", "port", cfg.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func runMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
