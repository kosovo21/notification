# Notification System

A high-performance, scalable notification system that sends messages across multiple platforms (SMS, WhatsApp, Telegram, Email) through a unified API interface.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen)

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Tech Stack](#-tech-stack)
- [Prerequisites](#-prerequisites)
- [Quick Start](#-quick-start)
- [API Documentation](#-api-documentation)
- [Configuration](#-configuration)
- [Development](#-development)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Monitoring](#-monitoring)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

### Core Capabilities
- 🚀 **Multi-Platform Support** - Send messages via SMS, WhatsApp, Telegram, and Email
- ⚡ **High Performance** - Handle 10,000+ messages per second
- 🔄 **Asynchronous Processing** - Queue-based architecture with RabbitMQ
- 📊 **Real-time Tracking** - Monitor delivery status for every message
- 🔐 **Secure** - API key authentication and rate limiting
- 📈 **Scalable** - Horizontal scaling with distributed workers
- 🔁 **Retry Mechanism** - Automatic retry with exponential backoff
- 📉 **Analytics** - Delivery rates, platform stats, and cost tracking

### Advanced Features
- **Priority Messaging** - High priority for OTP/critical messages
- **Bulk Sending** - Send to multiple recipients efficiently
- **Idempotency** - Prevent duplicate message sends
- **Rate Limiting** - Per-user/tier rate limits
- **Webhook Support** - Receive delivery status updates
- **Message Scheduling** - Send messages at specific times (optional)
- **Templates** - Reusable message templates with variables (optional)

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────────────────────────────────┐
│          API Gateway (Go)               │
│  ┌───────────────────────────────────┐  │
│  │ Auth │ Rate Limit │ Validation    │  │
│  └───────────────────────────────────┘  │
└──────┬──────────────────────┬───────────┘
       │                      │
       ▼                      ▼
┌──────────────┐      ┌──────────────┐
│  PostgreSQL  │      │  Redis Cache │
│   Messages   │      │ Rate Limits  │
└──────────────┘      └──────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│           RabbitMQ Exchange             │
│    ┌──────┬──────┬──────┬──────┐       │
│    │ SMS  │  WA  │ TG   │Email │       │
└────┴──────┴──────┴──────┴──────┴───────┘
       │      │      │      │
       ▼      ▼      ▼      ▼
┌─────────────────────────────────────────┐
│         Worker Services (Go)            │
│  ┌──────────────────────────────────┐   │
│  │   Platform Adapters              │   │
│  │  • Twilio    • WhatsApp Business │   │
│  │  • Telegram  • SendGrid          │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│    External Platform APIs               │
│  📱 SMS  💬 WhatsApp  ✈️ TG  📧 Email  │
└─────────────────────────────────────────┘
```

## 🛠️ Tech Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.21+ |
| **HTTP Framework** | Gin |
| **Database** | PostgreSQL 14+ |
| **Cache** | Redis 7+ |
| **Message Queue** | RabbitMQ 3.12+ |
| **Logging** | Zerolog |
| **Metrics** | Prometheus |
| **Visualization** | Grafana |
| **Container** | Docker |
| **Orchestration** | Kubernetes (optional) |

### Platform Integrations
- **SMS**: Twilio / Vonage / AWS SNS
- **Email**: SendGrid / AWS SES / Mailgun
- **WhatsApp**: WhatsApp Business API
- **Telegram**: Telegram Bot API

## 📦 Prerequisites

- Go 1.21 or higher
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 7+
- RabbitMQ 3.12+
- Platform accounts (Twilio, SendGrid, etc.)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/notification-system.git
cd notification-system
```

### 2. Start Infrastructure Services

```bash
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- Redis (port 6379)
- RabbitMQ (port 5672, Management UI: 15672)

### 3. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=notification_db
DB_USER=notification_user
DB_PASSWORD=your_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Platform Credentials
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_PHONE_NUMBER=+1234567890

SENDGRID_API_KEY=your_api_key
SENDGRID_FROM_EMAIL=noreply@yourdomain.com

WHATSAPP_API_KEY=your_api_key
WHATSAPP_PHONE_ID=your_phone_id

TELEGRAM_BOT_TOKEN=your_bot_token

# Server
SERVER_PORT=8080
LOG_LEVEL=info
```

### 4. Run Database Migrations

```bash
make migrate-up
# or
go run cmd/migrate/main.go up
```

### 5. Start API Gateway

```bash
make run-api
# or
go run cmd/server/main.go
```

API will be available at `http://localhost:8080`

### 6. Start Worker Service

```bash
make run-worker
# or
go run cmd/worker/main.go
```

### 7. Create Test User & API Key

```bash
make seed
# or
go run cmd/seed/main.go
```

This creates a test user with API key: `test-api-key-12345`

## 📚 API Documentation

### Authentication

All API requests require authentication via API Key:

```bash
curl -H "X-API-Key: your-api-key" https://api.example.com/api/v1/messages
```

### Base URL

```
Production: https://api.yourdomain.com/api/v1
Development: http://localhost:8080/api/v1
```

### Swagger UI (Interactive Docs)

Full interactive API documentation is available via Swagger UI once the server is running:

| Resource | URL |
|----------|-----|
| **Swagger UI** | [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) |
| **OpenAPI Spec (YAML)** | [http://localhost:8080/docs/swagger.yaml](http://localhost:8080/docs/swagger.yaml) |

The Swagger UI lets you explore all endpoints, view request/response schemas, and try out API calls directly from the browser.

### Endpoints Overview

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/health` | Health check | No |
| `GET` | `/version` | Build version info | No |
| `GET` | `/metrics` | Prometheus metrics | No |
| `POST` | `/api/v1/messages/send` | Send a message | ✅ |
| `POST` | `/api/v1/messages/bulk` | Bulk send messages | ✅ |
| `GET` | `/api/v1/messages/{id}` | Get message status | ✅ |
| `GET` | `/api/v1/messages` | List messages (paginated) | ✅ |
| `DELETE` | `/api/v1/messages/{id}` | Cancel a scheduled message | ✅ |
| `POST` | `/webhooks/twilio` | Twilio status callback | No |
| `POST` | `/webhooks/sendgrid` | SendGrid event callback | No |

### Rate Limits

Rate limits are applied per API key based on tier:

| Tier | Requests/Minute |
|------|----------------|
| Free | 60 |
| Basic | 300 |
| Premium | 1,000 |
| Enterprise | 10,000 |

## ⚙️ Configuration

### Configuration File

`config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  host: "localhost"
  port: 5432
  name: "notification_db"
  user: "notification_user"
  password: "${DB_PASSWORD}"
  max_open_conns: 25
  max_idle_conns: 5

redis:
  host: "localhost"
  port: 6379
  db: 0
  pool_size: 10

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  prefetch_count: 10

rate_limit:
  enabled: true
  tiers:
    free:
      requests_per_min: 60
    basic:
      requests_per_min: 300

platforms:
  sms:
    enabled: true
    provider: "twilio"
    rate_limit: 100
  whatsapp:
    enabled: true
    provider: "whatsapp_business"
    rate_limit: 80
  telegram:
    enabled: true
    rate_limit: 30
  email:
    enabled: true
    provider: "sendgrid"
    rate_limit: 200

logging:
  level: "info"
  format: "json"
```

### Environment Variables

Override config values using environment variables:

```bash
SERVER_PORT=8080
DB_HOST=localhost
REDIS_HOST=localhost
LOG_LEVEL=debug
```

## 💻 Development

### Project Structure

```
notification-system/
├── cmd/
│   ├── server/          # API Gateway entry point
│   ├── worker/          # Worker service entry point
│   ├── migrate/         # Database migration tool
│   └── seed/            # Database seeder
├── internal/
│   ├── adapter/         # Platform adapters (Twilio, SendGrid)
│   ├── auth/            # API key hashing & validation
│   ├── cache/           # Redis cache
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── metrics/         # Prometheus metric definitions
│   ├── middleware/      # Auth, rate limit, CORS, logging, metrics
│   ├── model/           # Data models & request/response types
│   ├── queue/           # RabbitMQ publisher/consumer
│   ├── repository/      # Database access layer
│   ├── router/          # Route definitions & Swagger UI
│   ├── scheduler/       # Scheduled message polling
│   ├── service/         # Business logic
│   └── worker/          # Worker logic
├── pkg/
│   └── logger/          # Logging utilities
├── docs/
│   ├── swagger.yaml     # OpenAPI 3.0 specification
│   └── docs.go          # Embed file for swagger.yaml
├── migrations/          # SQL migration files
├── docker/              # Dockerfiles (API & Worker)
├── k8s/                 # Kubernetes manifests
├── scripts/             # Utility scripts
├── .env.example
├── config.yaml
├── docker-compose.yml
├── Makefile
└── README.md
```

### Makefile Commands

```bash
# Development
make run-api              # Run API gateway
make run-worker           # Run worker service
make dev                  # Run with hot reload (air)

# Database
make migrate-up           # Run migrations
make migrate-down         # Rollback migrations
make seed                 # Seed database

# Testing
make test                 # Run unit tests
make test-integration     # Run integration tests
make test-coverage        # Generate coverage report
make lint                 # Run linter

# Build
make build                # Build binaries
make docker-build         # Build Docker images

# Cleanup
make clean                # Remove build artifacts
```

### Hot Reload (Development)

Install Air for hot reloading:

```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:

```bash
air
```

### Code Style

This project follows standard Go conventions:

- Run `gofmt` before committing
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `golangci-lint` for linting

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

## 🧪 Testing

### Run Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Integration tests (requires Docker)
make test-integration

# Specific package
go test ./internal/service/...
```

### Test Coverage

Current coverage: **85%**

View coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Load Testing

Using k6:

```bash
k6 run scripts/load-test.js
```

Expected performance:
- **Throughput**: 10,000 req/sec
- **Latency**: p95 < 100ms, p99 < 200ms
- **Error Rate**: < 0.1%

## 🏷️ Versioning

This project uses **semantic versioning** via git tags. Version info is injected at compile time using Go `ldflags`.

### Creating a Release

```bash
# Tag a release
git tag v1.0.0
git push origin v1.0.0
```

This triggers the Release workflow which builds and pushes Docker images to GHCR.

### Version Endpoint

The running server exposes version info at `GET /version`:

```json
{
  "version": "v1.0.0",
  "commit": "abc1234",
  "build_date": "2026-02-16T06:00:00Z"
}
```

### Local Build with Version

```bash
make build
# Output: Built v1.0.0 (abc1234) at 2026-02-16T06:00:00Z
```

## 🔄 CI/CD

The project includes two GitHub Actions workflows:

| Workflow | Trigger | Steps |
|----------|---------|-------|
| **CI** (`.github/workflows/ci.yml`) | Push to `main`, PRs | Lint → Test → Build |
| **Release** (`.github/workflows/release.yml`) | Tag push `v*` | Test → Build multi-arch images → Push to GHCR |

### Docker Images

On each tagged release, multi-platform images (`linux/amd64`, `linux/arm64`) are published to GitHub Container Registry:

```
ghcr.io/<owner>/notification-api:v1.0.0
ghcr.io/<owner>/notification-worker:v1.0.0
```

## 🚀 Deployment

### Docker

Build images locally with version info:

```bash
make docker-build
```

Run containers:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes

Deploy to Kubernetes:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/api-deployment.yaml
kubectl apply -f k8s/worker-deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

Scale workers:

```bash
kubectl scale deployment notification-worker --replicas=5
```

### Environment-Specific Configs

```bash
# Production
docker-compose -f docker-compose.prod.yml up -d

# Staging
docker-compose -f docker-compose.staging.yml up -d
```

## 📊 Monitoring

### Prometheus Metrics

Access metrics at: `http://localhost:8080/metrics`

**Key Metrics:**
- `api_requests_total` - Total API requests
- `api_request_duration_seconds` - Request latency
- `messages_published_total` - Messages published to queue
- `messages_processed_total` - Messages processed by workers
- `messages_delivered_total` - Successfully delivered messages
- `messages_failed_total` - Failed message deliveries
- `rate_limit_hits_total` - Rate limit hits

### Grafana Dashboards

Access Grafana: `http://localhost:3000`

**Default credentials:** admin/admin

**Available Dashboards:**
1. **API Performance** - Request rates, latency, errors
2. **Message Delivery** - Delivery rates by platform
3. **System Health** - CPU, memory, connections
4. **Business Metrics** - Platform usage, costs

### Logging

Structured JSON logs with fields:
- `request_id` - Unique request identifier
- `user_id` - User making request
- `duration` - Request duration
- `status` - HTTP status code
- `error` - Error message (if any)

View logs:

```bash
# API Gateway
docker logs -f notification-api

# Worker
docker logs -f notification-worker

# Filter by level
docker logs notification-api | grep '"level":"error"'
```

### Alerting

Configure alerts in `prometheus/alerts.yml`:

- API error rate > 1%
- Message delivery rate < 95%
- Worker queue depth > 10,000
- Database connection pool exhausted

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Code Review Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Code formatted (`gofmt`)
- [ ] Linter passes (`golangci-lint`)
- [ ] No security vulnerabilities
- [ ] Performance impact considered

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [RabbitMQ](https://www.rabbitmq.com/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)

## 📞 Support

- **Documentation**: [https://docs.yourdomain.com](https://docs.yourdomain.com)
- **Issues**: [GitHub Issues](https://github.com/yourusername/notification-system/issues)
- **Email**: support@yourdomain.com
- **Discord**: [Join our community](https://discord.gg/yourinvite)

## 🗺️ Roadmap

- [x] Message scheduling
- [x] Bulk message sending
- [x] Webhook status callbacks (Twilio & SendGrid)
- [x] Prometheus metrics & observability
- [x] Swagger / OpenAPI documentation
- [x] Kubernetes deployment manifests
- [ ] Message templates system
- [ ] Multi-language support
- [ ] A/B testing capability
- [ ] Advanced analytics dashboard
- [ ] Cost optimization engine
- [ ] Multi-tenancy support

---

**Built with ❤️ using Go**