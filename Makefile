PROJECT_NAME := lavoval

.PHONY: dev up down logs format format-web format-api lint lint-web lint-api test test-backend test-frontend migrate seed

dev:
	pnpm run dev:stack

up:
	docker compose --env-file .env.local up -d

down:
	docker compose --env-file .env.local down

logs:
	docker compose --env-file .env.local logs -f

## Format all code (frontend + backend)
format: format-web format-api

format-web:
	pnpm --dir apps/web format

## goimports = gofmt + organised import groups (stdlib / third-party / local).
## Install once: go install golang.org/x/tools/cmd/goimports@latest
format-api:
	cd apps/api && goimports -w -local github.com/sergehall/lavoval \
		$$(find . -name '*.go' -not -path './vendor/*')

## Lint all code (frontend + backend)
lint: lint-web lint-api

lint-web:
	pnpm --dir apps/web lint

## Requires golangci-lint. Install: https://golangci-lint.run/usage/install/
lint-api:
	cd apps/api && golangci-lint run ./...

test: test-backend test-frontend

test-backend:
	cd apps/api && go test ./...

test-frontend:
	pnpm --dir apps/web test --run

migrate:
	psql "$${DATABASE_URL}" -f infrastructure/db/migrations/001_init.sql
	psql "$${DATABASE_URL}" -f infrastructure/db/migrations/002_seed_admin.sql

seed:
	psql "$${DATABASE_URL}" -f infrastructure/db/seeds/001_demo_data.sql
