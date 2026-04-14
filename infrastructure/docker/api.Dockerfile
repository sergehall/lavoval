FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY backend/api/go.mod backend/api/go.sum ./backend/api/
RUN cd backend/api && go mod download
COPY backend/api ./backend/api
RUN cd backend/api && CGO_ENABLED=0 GOOS=linux go build -o /bin/lavoval-api ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /bin/lavoval-api /app/server
EXPOSE 8080
CMD ["/app/server"]
