.PHONY: run test migrate swagger fmt lint

run:
	go run ./cmd/server

test:
	go test ./... -v -count=1

migrate:
	go run ./cmd/server -migrate-only

swagger:
	swag init -g cmd/server/main.go -o docs

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...
