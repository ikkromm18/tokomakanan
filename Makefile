.PHONY: all build run test test-coverage migrate rollback clean docker-up docker-down

APP_NAME=tokomakanan
BINARY=bin/api

all: test build

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BINARY) cmd/api/main.go

run:
	@go run cmd/api/main.go

test:
	@go test -v ./...

test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

migrate:
	@go run cmd/api/main.go --migrate

rollback:
	@go run cmd/api/main.go --rollback

clean:
	@rm -rf bin/ coverage.out

docker-up:
	@docker compose up -d

docker-down:
	@docker compose down
