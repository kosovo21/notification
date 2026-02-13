# Notification System

A high-performance, scalable notification system that sends messages across multiple platforms (SMS, WhatsApp, Telegram, Email) through a unified API interface.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
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

### Endpoints

#### Send Message

Send a single message to one or more recipients.

**Request:**
```http
POST /api/v1/messages/send
Content-Type: application/json
X-API-Key: your-api-key

{
  "subject": "Welcome to Our Service",
  "message": "Hello John, thanks for signing up!",
  "from": "YourApp",
  "to": ["+1234567890", "+0987654321"],
  "platform": "sms",
  "priority": 1
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "7289c6c3-3bba-40b0-b991-c4a8df51c495",
  "recipients_count": 2,
  "estimated_delivery": "2024-01-20T10:25:30Z",
  "request_id": "req_abc123"
}
```

**Parameters:**
- `subject` (string, required): Message subject/title (max 200 chars)
- `message` (string, required): Message body (max 5000 chars)
- `from` (string, required): Sender identifier (max 100 chars)
- `to` (array, required): List of recipients (1-1000)
- `platform` (string, required): `sms`, `whatsapp`, `telegram`, or `email`
- `priority` (int, optional): 0=low, 1=normal (default), 2=high

#### Get Message Status

Retrieve the delivery status of a message.

**Request:**
```http
GET /api/v1/messages/{message_id}
X-API-Key: your-api-key
```

**Response:**
```json
{
  "success": true,
  "message_id": "7289c6c3-3bba-40b0-b991-c4a8df51c495",
  "status": {
    "message_id": "7289c6c3-3bba-40b0-b991-c4a8df51c495",
    "subject": "Welcome to Our Service",
    "platform": "sms",
    "total_recipients": 2,
    "summary": {
      "queued": 0,
      "processing": 0,
      "sent": 1,
      "delivered": 1,
      "failed": 0,
      "pending": 0
    },
    "recipients": [
      {
        "recipient": "+1234****90",
        "status": 3,
        "delivered_at": "2024-01-20T10:25:45Z"
      },
      {
        "recipient": "+0987****21",
        "status": 2,
        "sent_at": "2024-01-20T10:25:30Z"
      }
    ],
    "created_at": "2024-01-20T10:25:00Z"
  }
}
```

**Status Codes:**
- `0` - Queued
- `1` - Processing
- `2` - Sent
- `3` - Delivered
- `4` - Failed
- `5` - Pending (will retry)
- `6` - Cancelled

#### List Messages

Get paginated list of messages with filters.

**Request:**
```http
GET /api/v1/messages?page=1&limit=20&platform=sms&status=3
X-API-Key: your-api-key
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (max: 100, default: 20)
- `platform` (string): Filter by platform
- `status` (int): Filter by status
- `from` (timestamp): Filter from date
- `to` (timestamp): Filter to date

**Response:**
```json
{
  "success": true,
  "messages": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### Send Bulk Messages

Send multiple different messages in one request.

**Request:**
```http
POST /api/v1/messages/send/bulk
X-API-Key: your-api-key

{
  "messages": [
    {
      "subject": "Order Confirmation",
      "message": "Your order #123 has been confirmed",
      "from": "Shop",
      "to": ["+1234567890"],
      "platform": "sms"
    },
    {
      "subject": "Welcome Email",
      "message": "Welcome to our platform!",
      "from": "support@example.com",
      "to": ["user@email.com"],
      "platform": "email"
    }
  ]
}
```

#### Retry Failed Message

Retry sending a failed message.

**Request:**
```http
POST /api/v1/messages/{message_id}/retry
X-API-Key: your-api-key
```

#### Cancel Scheduled Message

Cancel a scheduled message before it's sent.

**Request:**
```http
DELETE /api/v1/messages/{message_id}
X-API-Key: your-api-key
```

### Rate Limits

Rate limits are applied per API key based on tier:

| Tier | Requests/Minute | Burst |
|------|----------------|-------|
| Free | 60 | 10 |
| Basic | 300 | 50 |
| Premium | 1,000 | 100 |
| Enterprise | 10,000 | 500 |

**Rate Limit Headers:**
```
X-RateLimit-Limit: 300
X-RateLimit-Remaining: 245
X-RateLimit-Reset: 1674234567
```

### Error Responses

All errors follow this format:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "fields": {
      "platform": "must be one of: sms, whatsapp, telegram, email"
    }
  }
}
```

**Common Error Codes:**
- `UNAUTHORIZED` - Invalid or missing API key
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `VALIDATION_ERROR` - Invalid request data
- `NOT_FOUND` - Message not found
- `PLATFORM_NOT_ALLOWED` - Platform not enabled for account
- `INTERNAL_ERROR` - Server error

### Swagger Documentation

Interactive API documentation available at:
```
http://localhost:8080/swagger/index.html
```

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
│   ├── config/          # Configuration management
│   ├── middleware/      # HTTP middleware
│   ├── handler/         # HTTP handlers
│   ├── service/         # Business logic
│   ├── repository/      # Database access
│   ├── model/           # Data models
│   ├── queue/           # RabbitMQ publisher/consumer
│   ├── cache/           # Redis cache
│   ├── auth/            # Authentication
│   ├── adapter/         # Platform adapters
│   ├── worker/          # Worker logic
│   └── router/          # Route definitions
├── pkg/
│   ├── logger/          # Logging utilities
│   ├── validator/       # Custom validators
│   └── errors/          # Error definitions
├── migrations/          # SQL migrations
├── docker/              # Docker files
├── docs/                # Documentation
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

## 🚀 Deployment

### Docker

Build images:

```bash
docker build -t notification-api:latest -f docker/api.Dockerfile .
docker build -t notification-worker:latest -f docker/worker.Dockerfile .
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

- [ ] Message templates system
- [ ] Multi-language support
- [ ] A/B testing capability
- [ ] Advanced analytics dashboard
- [ ] Webhook retry mechanism
- [ ] Message scheduling
- [ ] Cost optimization engine
- [ ] Multi-tenancy support

---

**Built with ❤️ using Go**