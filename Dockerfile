# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache bash ca-certificates postgresql-client
WORKDIR /app
COPY --from=builder /build/server ./
COPY scripts/ ./scripts/
COPY infrastructure/db/migrations/ ./infrastructure/db/migrations/
CMD ["./server"]
