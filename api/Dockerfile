FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /signal \
    ./cmd/api/main.go

# Final minimal image
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /signal /app/signal

# Copy migrations
COPY db/migrations /app/db/migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/signal"]
