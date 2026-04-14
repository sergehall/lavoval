FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
RUN cd apps/api && go mod download
COPY apps/api ./apps/api
RUN cd apps/api && CGO_ENABLED=0 GOOS=linux go build -o /bin/lavoval-api ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /bin/lavoval-api /app/server
EXPOSE 8080
CMD ["/app/server"]
