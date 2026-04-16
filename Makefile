SHELL := /bin/bash

.PHONY: \
	dev \
	up \
	down \
	logs \
	format \
	format-web \
	format-api \
	lint \
	lint-web \
	lint-api \
	test \
	test-backend \
	test-frontend \
	migrate \
	migrate-prod \
	migrate-status \
	seed

dev:
	pnpm run dev:stack

up:
	pnpm run infra:up

down:
	pnpm run infra:down

logs:
	pnpm run infra:logs

format: format-web format-api

format-web:
	pnpm run format:web

format-api:
	pnpm run format:api

lint: lint-web lint-api

lint-web:
	pnpm run lint:web

lint-api:
	pnpm run lint:api

test: test-backend test-frontend

test-backend:
	pnpm run test:api

test-frontend:
	pnpm run test:web

migrate:
	pnpm run db:migrate

migrate-prod:
	pnpm run db:migrate:prod

migrate-status:
	pnpm run db:migrate:status

seed:
	pnpm run db:seed
