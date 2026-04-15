# Lavoval

Lavoval is a fullstack foundation for building a skill-exchange marketplace for the AI era. The product idea is simple: users do not just consume content, they publish skills, discover other people's expertise, and exchange practical knowledge in a world where AI accelerates creation but human context, judgment, and lived experience still matter.

The repository is structured as a real product foundation: public marketplace area, authentication and authorization, personal account area, admin panel, Go REST API, PostgreSQL persistence, tests, Docker-backed local infrastructure, and growth-oriented architecture.

## Stack

- Frontend: Next.js App Router, React 19, TypeScript, server actions
- Backend: Go 1.26.x, Chi router, JWT auth, pgx, layered architecture
- Database: PostgreSQL 17
- Tooling: pnpm workspaces, Vitest, go test, Prettier, Docker Compose, Makefile

## Product direction

- Lavoval is centered on peer-to-peer skill exchange between users.
- Users can create, publish, and evolve skill pages that package expertise into reusable modules.
- The public catalog acts as the discovery layer for finding people, topics, and practical know-how.
- Personal account flows support identity, authorship, and future exchange history.
- Admin flows provide governance for quality, moderation, and marketplace health.

In AI-heavy workflows, the app positions human skill as something discoverable, exchangeable, and improvable instead of something hidden in resumes, chats, or scattered docs.

## Architecture summary

- `apps/web` owns marketplace presentation, route protection, server actions, reusable UI primitives, and typed API calls.
- `apps/api` owns HTTP transport, validation, auth, business logic, and database access.
- `packages/contracts` defines frontend-safe contracts and schemas for predictable API usage.
- `infrastructure` contains Docker assets, migrations, and seeds.
- `docs` documents architecture and API boundaries.

A fuller explanation lives in [docs/architecture.md](./docs/architecture.md).

## Project structure

```text
lavoval/
├── apps/web
├── apps/api
├── packages/contracts
├── infrastructure/docker
├── infrastructure/db/migrations
├── infrastructure/db/seeds
├── docs
├── tests
├── .env.example
├── docker-compose.yml
├── Makefile
└── README.md
```

## Local setup

1. Copy environment variables:
   - local infrastructure/dev setup already has `.env.local`
   - if you want a separate shared env file, copy `.env.example` as needed
2. Install dependencies:
   - `pnpm install`
3. Start the full local stack:
   - `pnpm run dev:stack`
4. Stop the full local stack safely:
   - `pnpm run dev:stack:stop`
5. Open the apps:
   - Frontend: `http://localhost:3000`
   - API: `http://localhost:8080`
   - Health check: `http://localhost:8080/healthz`

## CLI

Lavoval now has an early `v2` CLI surface powered by the shared SDK.

- `pnpm run lavoval -- help`
- `pnpm run lavoval -- dev`
- `pnpm run lavoval -- auth login --email admin@lavoval.local --password ChangeMe123!`
- `pnpm run lavoval -- auth me`
- `pnpm run lavoval -- auth whoami`
- `pnpm run lavoval -- auth logout`
- `pnpm run lavoval -- skills list`
- `pnpm run lavoval -- skills search prompt --status published --creator serge --entrypoint text-summary-mock`
- `pnpm run lavoval -- skills get <skill-id>`
- `pnpm run lavoval -- runs list --token <access-token>`
- `pnpm run lavoval -- runs get <run-id>`
- `pnpm run lavoval -- runs replay <run-id> --text "Retry with a tighter input" --config '{"limit":3}' --as-json`
- `pnpm run lavoval -- admin runs list`
- `pnpm run lavoval -- admin runs failures`
- `pnpm run lavoval -- admin runs stats --entrypoint echo --creator serge`
- `pnpm run lavoval -- admin runs get <run-id>`
- `pnpm run lavoval -- run <skill-id> --text "Hello runtime" --token <access-token>`

Environment variables for CLI usage:

- `LAVOVAL_API_URL` defaults to `http://localhost:8080`
- `LAVOVAL_ACCESS_TOKEN` can be used instead of passing `--token`
- CLI also persists a local session in `.lavoval/session.json` after `auth login`

## Environment variables

See [.env.example](./.env.example) for the full list.

Important values:

- `APP_URL`: frontend base URL used to build verification links
- `NEXT_PUBLIC_API_URL`: backend base URL for the frontend
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: signing secret for access and refresh tokens
- `JWT_ACCESS_TTL` and `JWT_REFRESH_TTL`: token lifetimes
- `ADMIN_SEED_EMAIL`, `ADMIN_SEED_PASSWORD`: seed account defaults
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM_EMAIL`: Google SMTP delivery settings for confirmation emails

Google email confirmation setup:

- Enable 2-Step Verification on the Google account you want to send from
- Create a Google App Password and place it in `SMTP_PASSWORD`
- Keep `SMTP_HOST=smtp.gmail.com` and `SMTP_PORT=587`
- Make sure `APP_URL` points at the frontend domain users will open from their inbox

For local infrastructure runs, `docker-compose.yml` uses `.env.local` and starts only PostgreSQL.

## Database and migrations

SQL migrations live in `infrastructure/db/migrations`.

- `001_init.sql`: core schema for users, profiles, skills, modules, and enrollments
- `002_seed_admin.sql`: default admin account for local development
- `001_demo_data.sql`: sample user, published skill, and module records

Run migrations manually when needed:

```bash
pnpm run db:migrate
pnpm run db:migrate:status
pnpm run db:seed
```

Professional local migration flow:

- `db:migrate` applies only new SQL files and records them in `schema_migrations`
- `db:migrate:status` shows which migration files are already applied
- `dev:stack` automatically runs `db:migrate` after infrastructure startup
- seed files remain manual so demo data is an explicit choice

## Frontend areas

- Public: marketplace landing page, login, registration, email confirmation, recovery placeholder, skills catalog
- Account: dashboard, profile, authored skills, skill detail, future exchange activity
- Admin: dashboard, users list, skills CRUD and governance surface

Protection is enforced in two layers:

- `apps/web/src/proxy.ts`
- server-side account/admin layouts

## Backend design

The API uses explicit layers:

- Handlers: HTTP concerns, decoding, validation, status codes
- Services: use cases and business behavior
- Repositories: PostgreSQL access only
- Middleware: auth, RBAC, request id, recovery, logging
- Domain models: framework-light entities and enums

Primary endpoints are documented in [docs/api-overview.md](./docs/api-overview.md).

## Roles and access model

- `user`: authenticated account area and visible skills catalog
- `admin`: all user capabilities plus admin dashboard, users list, and skill CRUD

The structure is intentionally ready for:

- richer permission matrices
- teams and organizations
- moderation workflows
- audit logs and analytics
- background jobs and notifications

## Tests

Frontend tests:

```bash
pnpm --dir apps/web test
```

Backend tests:

```bash
cd apps/api && go test ./...
```

Included examples cover:

- auth service email verification and token issuance
- skill service create flow
- auth handler validation failure path
- reusable button component rendering
- protected route design via `proxy.ts` and account/admin server guards

## Formatting and quality

```bash
pnpm run format
pnpm run lint
pnpm run test
```

- Frontend formatting uses `prettier --write`
- Backend formatting uses `gofmt`
- Tests are split between Vitest and `go test`

## Docker workflow

```bash
pnpm run infra:up
pnpm run infra:down
pnpm run infra:logs
```

Compose service:

- `postgres`

`make down` and `pnpm run infra:down` use a safe stop flow and do not remove the local container volume data.

## Extension points

The repository is shaped to support the next product steps without a painful rewrite:

- peer-to-peer exchange mechanics and matching
- reputation, trust, and contributor history
- permissions matrix and policy engine
- billing and subscriptions
- notifications and email workflows
- file uploads and media management
- analytics and audit logs
- search, filters, and pagination
- skill versioning and moderation
- background jobs and task orchestration

## Notes

- The frontend uses server actions as a pragmatic auth boundary so tokens can stay in HTTP-only cookies.
- The backend exposes stateless JWT auth with email confirmation now, while leaving room for future refresh rotation, stronger session tracking, rate limiting, and audit logging.
- The shared contracts package currently serves TypeScript consumers. A future OpenAPI-driven workflow can become the cross-language contract source if the product needs stronger generation flows.
