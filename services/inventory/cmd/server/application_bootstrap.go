package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/rts/inventory/docs"

	"github.com/rts/inventory/internal/cache"
	"github.com/rts/inventory/internal/config"
	"github.com/rts/inventory/internal/handler"
	"github.com/rts/inventory/internal/middleware"
	"github.com/rts/inventory/internal/publisher"
	"github.com/rts/inventory/internal/repository"
	"github.com/rts/inventory/internal/router"
	"github.com/rts/inventory/internal/service"
)

type Repositories struct {
	inventory   *repository.InventoryRepository
	warehouse   *repository.WarehouseRepository
	stock       *repository.StockRepository
	movement    *repository.MovementRepository
	reservation *repository.ReservationRepository
}

type Services struct {
	inventory    *service.InventoryService
	movement     *service.MovementService
	reservation  *service.ReservationService
	availability *service.AvailabilityService
}

type Handlers struct {
	inventory    *handler.InventoryHandler
	warehouse    *handler.WarehouseHandler
	movement     *handler.MovementHandler
	reservation  *handler.ReservationHandler
	availability *handler.AvailabilityHandler
}

type Application struct {
	handler            http.Handler
	inventoryService   *service.InventoryService
	reservationService *service.ReservationService
}

func initializeApplication(
	cfg *config.Config,
	pool *pgxpool.Pool,
	redisCache *cache.RedisCache,
	eventPublisher *publisher.EventPublisher,
) *Application {
	repositories := initializeRepositories(pool)
	services := initializeServices(repositories, redisCache, eventPublisher)
	handlers := initializeHandlers(services)

	return &Application{
		handler:            initializeHTTPHandler(cfg, handlers),
		inventoryService:   services.inventory,
		reservationService: services.reservation,
	}
}

func initializeRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		inventory:   repository.NewInventoryRepository(pool),
		warehouse:   repository.NewWarehouseRepository(pool),
		stock:       repository.NewStockRepository(pool),
		movement:    repository.NewMovementRepository(pool),
		reservation: repository.NewReservationRepository(pool),
	}
}

func initializeServices(
	repositories *Repositories,
	redisCache *cache.RedisCache,
	eventPublisher *publisher.EventPublisher,
) *Services {
	inventoryService := service.NewInventoryService(
		repositories.inventory,
		repositories.stock,
		repositories.warehouse,
		repositories.reservation,
		eventPublisher,
	)

	movementService := service.NewMovementService(
		repositories.movement,
		repositories.stock,
		repositories.inventory,
		redisCache,
		eventPublisher,
	)

	reservationService := service.NewReservationService(
		repositories.reservation,
		repositories.stock,
		repositories.inventory,
		repositories.movement,
		redisCache,
		eventPublisher,
	)

	availabilityService := service.NewAvailabilityService(
		repositories.stock,
		redisCache,
	)

	return &Services{
		inventory:    inventoryService,
		movement:     movementService,
		reservation:  reservationService,
		availability: availabilityService,
	}
}

func initializeHandlers(services *Services) *Handlers {
	return &Handlers{
		inventory:    handler.NewInventoryHandler(services.inventory),
		warehouse:    handler.NewWarehouseHandler(services.inventory),
		movement:     handler.NewMovementHandler(services.movement),
		reservation:  handler.NewReservationHandler(services.reservation),
		availability: handler.NewAvailabilityHandler(services.availability),
	}
}

func initializeHTTPHandler(cfg *config.Config, handlers *Handlers) http.Handler {
	apiHandler := router.New(router.Props{
		JWTSecret:           cfg.JWTAccessSecret,
		InventoryHandler:    handlers.inventory,
		WarehouseHandler:    handlers.warehouse,
		MovementHandler:     handlers.movement,
		ReservationHandler:  handlers.reservation,
		AvailabilityHandler: handlers.availability,
	})

	swaggerMux := http.NewServeMux()
	swaggerMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))
	swaggerProtected := middleware.SwaggerBasicAuth(cfg.SwaggerUsername, cfg.SwaggerPassword)(swaggerMux)

	combinedMux := http.NewServeMux()
	combinedMux.Handle("/swagger/", swaggerProtected)
	combinedMux.Handle("/", apiHandler)

	return combinedMux
}
