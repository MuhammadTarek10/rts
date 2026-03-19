package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             int
	DatabaseURL      string
	RabbitMQURI      string
	RedisURL         string
	JWTAccessSecret  string
	RabbitMQExchange string
	RabbitMQQueue    string
	CatalogQueueName string
	SwaggerUsername  string
	SwaggerPassword  string
	MigrationsPath   string
}

func Load() (*Config, error) {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		port = parsed
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	rabbitURI := os.Getenv("RABBITMQ_URI")
	if rabbitURI == "" {
		return nil, fmt.Errorf("RABBITMQ_URI is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_ACCESS_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	exchange := os.Getenv("RABBITMQ_EXCHANGE_NAME")
	if exchange == "" {
		exchange = "inventory.exchange"
	}

	queue := os.Getenv("RABBITMQ_QUEUE_NAME")
	if queue == "" {
		queue = "inventory.events"
	}

	catalogQueue := os.Getenv("CATALOG_QUEUE_NAME")
	if catalogQueue == "" {
		catalogQueue = "inventory.catalog-events"
	}

	swaggerUser := os.Getenv("SWAGGER_USERNAME")
	if swaggerUser == "" {
		swaggerUser = "admin"
	}

	swaggerPass := os.Getenv("SWAGGER_PASSWORD")
	if swaggerPass == "" {
		swaggerPass = "admin"
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	return &Config{
		Port:             port,
		DatabaseURL:      dbURL,
		RabbitMQURI:      rabbitURI,
		RedisURL:         redisURL,
		JWTAccessSecret:  jwtSecret,
		RabbitMQExchange: exchange,
		RabbitMQQueue:    queue,
		CatalogQueueName: catalogQueue,
		SwaggerUsername:  swaggerUser,
		SwaggerPassword:  swaggerPass,
		MigrationsPath:   migrationsPath,
	}, nil
}
