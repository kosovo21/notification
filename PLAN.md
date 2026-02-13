# Implementation Plan

This plan breaks down the development of the Notification System into executable steps for code generation.

## Phase 1: Foundation & Infrastructure Setup

### Step 1.1: Project Initialization
- Initialize Go module: `go mod init notification-system`
- Create standard directory structure:
  - `cmd/server`, `cmd/worker`, `cmd/migrate`, `cmd/seed`
  - `internal/config`, `internal/model`, `internal/repository`, `internal/service`, `internal/handler`, `internal/middleware`
  - `pkg/logger`, `pkg/utils`
- Setup `.gitignore` for Go.

### Step 1.2: Configuration Management
- Implement `internal/config`:
  - Define `Config` struct matching `README.md` specs (DB, Redis, RabbitMQ, RateLimits).
  - Use `viper` to load from `config.yaml` and environment variables.
- Create `config.yaml` template.
- Create `.env.example`.

### Step 1.3: Docker Environment
- Create `docker-compose.yml`:
  - PostgreSQL 14+
  - Redis 7+
  - RabbitMQ 3.12+ (with Management Plugin)
- Configure ports and environment variables for local dev.

### Step 1.4: Database Schema & Migrations
- Setup `golang-migrate`.
- Create SQL migrations (`migrations/`):
  - `001_create_users.sql`: `users` table (id, email, api_key, role, rate_limit_tier).
  - `002_create_messages.sql`: `messages` table (id, user_id, content, status, scheduled_at, created_at).
  - `003_create_recipients.sql`: `message_recipients` table (id, message_id, recipient, platform, status, provider_id).
- Create `cmd/migrate/main.go` tool to run migrations.

---

## Phase 2: Core Domain & Data Layer

### Step 2.1: Domain Models
- Define structs in `internal/model`:
  - `User`, `Message`, `Recipient`
  - `CreateMessageRequest`, `MessageResponse`
- Add struct tags for JSON and DB.

### Step 2.2: Repository Layer
- Implement `internal/repository`:
  - `UserRepository`: `GetByAPIKey`, `GetByID`.
  - `MessageRepository`: `Create` (transactional), `UpdateStatus`, `GetByID`, `List`.
  - `RecipientRepository`: `BatchCreate`, `UpdateStatus`.
- Use `sqlx` or `gorm` (as per preference, standard `sql/database` + `sqlx` recommended for performance).

---

## Phase 3: API Gateway - Core

### Step 3.1: HTTP Server Setup
- Initialize Gin engine in `internal/router`.
- Create `cmd/server/main.go`.
- Implement `internal/middleware`:
  - `Recovery`, `Logger` (using `zerolog`).
  - `CORS`.

### Step 3.2: Authentication & Rate Limiting
- Implement `AuthMiddleware`:
  - Validate `X-API-Key` header against DB/Redis.
- Implement `RateLimitMiddleware` (Tier-based):
  - Use Redis to track request counts.
  - Tiers: Free (60/min), Basic (300/min), Premium (1000/min).
  - Return `429` if exceeded.

### Step 3.3: Message Handlers (Basic)
- Implement `internal/handler/message_handler.go`:
  - `SendMessage`: Parse request, validate, call Service.
  - `GetMessageStatus`: key-based lookup.
  - `ListMessages`: Pagination + Filters.

---

## Phase 4: Service Layer & Queue Integration

### Step 4.1: RabbitMQ Infrastructure
- Implement `internal/queue`:
  - `Publisher`: Connect to RabbitMQ, declare Exchange (`notification.exchange`).
  - `Consumer`: Base consumer logic.
- Define Routing Keys: `sms`, `email`, `whatsapp`, `telegram`.

### Step 4.2: Message Service Logic
- Implement `internal/service/message_service.go`:
  - `Send`: 
    1. Validate Request.
    2. Persist Message & Recipients to DB (Status: PENDING).
    3. **Fan-out**: Publish individual events to RabbitMQ for each recipient.
    4. Update DB Status (Status: QUEUED).

---

## Phase 5: Worker Service

### Step 5.1: Worker Setup
- Create `cmd/worker/main.go`.
- Initialize RabbitMQ Consumer.
- Listen to queues: `notification.sms`, `notification.email`, etc.

### Step 5.2: Platform Adapters (Interfaces)
- Define `Sender` interface in `internal/adapter`:
  - `Send(ctx, recipient, content) (providerID, error)`

### Step 5.3: Adapter Implementations
- **SMS**: Implement `TwilioAdapter`.
- **Email**: Implement `SendGridAdapter`.
- **Mock**: Implement `MockAdapter` for testing/local dev.

### Step 5.4: Worker Message Processing
- Implement `Worker`:
  - Receive message.
  - Select Adapter based on routing key/type.
  - Call `adapter.Send()`.
  - Update `message_recipients` status (SENT/FAILED).
  - Handle Retries (Ack/Nack).

---

## Phase 6: Advanced Features

### Step 6.1: Scheduler
- Implement `internal/scheduler`:
  - Polling loop (Ticker).
  - `ScanScheduledMessages`: query `messages` where `scheduled_at <= now` AND `status = 'scheduled'`.
  - Publish to RabbitMQ.
  - Update status to `queued`.
- Integrate into `cmd/worker` or separate service.

### Step 6.2: Bulk Sending API
- Implement `POST /api/v1/messages/bulk`.
- Logic: Fan-out strategy (reuse Service logic to split into individual queue messages).

### Step 6.3: Cancellation
- Implement `DELETE /api/v1/messages/{id}`.
- Logic: Check if status is `scheduled`. If yes, update to `cancelled`.

---

## Phase 7: Webhooks & Tracking

### Step 7.1: Webhook Handlers
- Create `internal/handler/webhook_handler.go`.
- Endpoints: `/webhooks/twilio`, `/webhooks/sendgrid`.
- Update `message_recipients` based on provider callbacks (DELIVERED, READ, FAILED).

---

## Phase 8: Observability & Deployment

### Step 8.1: Metrics
- Integrate `prometheus/client_golang` in Middleware and Worker.
- Metrics: `http_requests_total`, `messages_published`, `messages_processed`.

### Step 8.2: Deployment Configs
- Create `Dockerfile.api` and `Dockerfile.worker`.
- Create `k8s/` manifests (Deployment, Service, Secret).
