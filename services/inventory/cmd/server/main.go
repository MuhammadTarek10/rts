package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/rts/inventory/docs"

	"github.com/rts/inventory/internal/cache"
	"github.com/rts/inventory/internal/config"
	"github.com/rts/inventory/internal/consumer"
	"github.com/rts/inventory/internal/handler"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/publisher"
	"github.com/rts/inventory/internal/repository"
	"github.com/rts/inventory/internal/router"
	"github.com/rts/inventory/internal/service"
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

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

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

	// Database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	// Redis cache
	redisCache := cache.NewRedisCache(cfg.RedisURL)
	if err := redisCache.Ping(ctx); err != nil {
		slog.Warn("redis not available, running without cache", "error", err)
	} else {
		slog.Info("connected to redis")
	}
	defer redisCache.Close()

	// Event publisher
	eventPublisher := publisher.NewEventPublisher(cfg.RabbitMQURI, cfg.RabbitMQExchange, cfg.RabbitMQQueue)
	if err := eventPublisher.Connect(); err != nil {
		slog.Warn("failed to connect event publisher to RabbitMQ", "error", err)
	} else {
		slog.Info("event publisher connected to RabbitMQ")
	}
	defer eventPublisher.Close()

	// Repositories
	inventoryRepo := repository.NewInventoryRepository(pool)
	warehouseRepo := repository.NewWarehouseRepository(pool)
	stockRepo := repository.NewStockRepository(pool)
	movementRepo := repository.NewMovementRepository(pool)
	reservationRepo := repository.NewReservationRepository(pool)

	// Services
	inventorySvc := service.NewInventoryService(inventoryRepo, stockRepo, warehouseRepo, reservationRepo, eventPublisher)
	movementSvc := service.NewMovementService(movementRepo, stockRepo, inventoryRepo, redisCache, eventPublisher)
	reservationSvc := service.NewReservationService(reservationRepo, stockRepo, inventoryRepo, movementRepo, redisCache, eventPublisher)
	availabilitySvc := service.NewAvailabilityService(stockRepo, redisCache)

	// Handlers
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)
	warehouseHandler := handler.NewWarehouseHandler(inventorySvc)
	movementHandler := handler.NewMovementHandler(movementSvc)
	reservationHandler := handler.NewReservationHandler(reservationSvc)
	availabilityHandler := handler.NewAvailabilityHandler(availabilitySvc)

	// Router
	mux := router.New(
		cfg.JWTAccessSecret,
		cfg.SwaggerUsername,
		cfg.SwaggerPassword,
		inventoryHandler,
		warehouseHandler,
		movementHandler,
		reservationHandler,
		availabilityHandler,
	)

	// Swagger UI (protected by basic auth)
	swaggerMux := http.NewServeMux()
	swaggerMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))
	swaggerProtected := middleware.SwaggerBasicAuth(cfg.SwaggerUsername, cfg.SwaggerPassword)(swaggerMux)

	// Combine main router with swagger
	combinedMux := http.NewServeMux()
	combinedMux.Handle("/swagger/", swaggerProtected)
	combinedMux.Handle("/", mux)

	// Catalog event consumer
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()

	catalogConsumer := consumer.NewCatalogConsumer(cfg.RabbitMQURI, cfg.CatalogQueueName, inventorySvc)
	if err := catalogConsumer.Start(consumerCtx); err != nil {
		slog.Warn("failed to start catalog consumer", "error", err)
	} else {
		slog.Info("catalog consumer started")
	}
	defer catalogConsumer.Close()

	// Reservation expiry sweeper
	sweeperCtx, sweeperCancel := context.WithCancel(ctx)
	defer sweeperCancel()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-sweeperCtx.Done():
				slog.Info("reservation sweeper stopped")
				return
			case <-ticker.C:
				processed, err := reservationSvc.ExpireBatch(sweeperCtx, 100)
				if err != nil {
					slog.Error("reservation sweeper error", "error", err)
				} else if processed > 0 {
					slog.Info("reservation sweeper processed expired reservations", "count", processed)
				}
			}
		}
	}()
	slog.Info("reservation expiry sweeper started")

	// HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      combinedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
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
	}()

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
