.PHONY: run-api run-worker dev migrate-up migrate-down seed test test-integration test-coverage lint build docker-build clean

# Version info — extracted from git
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     = -s -w \
              -X notification-system/internal/version.Version=$(VERSION) \
              -X notification-system/internal/version.Commit=$(COMMIT) \
              -X notification-system/internal/version.BuildDate=$(BUILD_DATE)

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
	go build -ldflags '$(LDFLAGS)' -o bin/server cmd/server/main.go
	go build -ldflags '$(LDFLAGS)' -o bin/worker cmd/worker/main.go
	go build -o bin/migrate cmd/migrate/main.go
	@echo "Built $(VERSION) ($(COMMIT)) at $(BUILD_DATE)"

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t notification-api:$(VERSION) \
		-t notification-api:latest \
		-f docker/api.Dockerfile .
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t notification-worker:$(VERSION) \
		-t notification-worker:latest \
		-f docker/worker.Dockerfile .

# Cleanup
clean:
	rm -rf bin/ coverage.out coverage.html
