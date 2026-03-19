.PHONY: run test migrate swagger fmt lint

SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.4

run:
	go run ./cmd/server

test:
	go test ./... -v -count=1

migrate:
	go run ./cmd/server -migrate-only

swagger:
	$(SWAG) init -g cmd/server/main.go -o docs

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...
