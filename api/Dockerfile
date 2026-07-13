FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Copy go mod files first — better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /signal \
    ./cmd/api/main.go

# Minimal final image
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=builder /signal /app/signal
COPY --from=builder /app/db/migrations /app/db/migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["/app/signal"]
