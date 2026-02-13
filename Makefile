.PHONY: run-api run-worker dev migrate-up migrate-down seed test test-integration test-coverage lint build docker-build clean

# Development
run-api:
	go run cmd/server/main.go

run-worker:
	go run cmd/worker/main.go

dev:
	air

# Database
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go -direction down

seed:
	go run cmd/seed/main.go

# Testing
test:
	go test ./... -v -short

test-integration:
	go test ./... -v -run Integration

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

# Build
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/worker cmd/worker/main.go
	go build -o bin/migrate cmd/migrate/main.go

docker-build:
	docker build -t notification-api:latest -f docker/api.Dockerfile .
	docker build -t notification-worker:latest -f docker/worker.Dockerfile .

# Cleanup
clean:
	rm -rf bin/ coverage.out coverage.html
