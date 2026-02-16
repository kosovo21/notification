# ---------- builder ----------
FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
    -X notification-system/internal/version.Version=${VERSION} \
    -X notification-system/internal/version.Commit=${COMMIT} \
    -X notification-system/internal/version.BuildDate=${BUILD_DATE}" \
    -o /worker cmd/worker/main.go

# ---------- runtime ----------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /worker /worker
COPY config.yaml /config.yaml

ENTRYPOINT ["/worker"]
