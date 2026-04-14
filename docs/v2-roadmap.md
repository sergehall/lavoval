# Lavoval V1/V2 Roadmap

## Purpose

This document fixes the current product state as `v1` and defines the large refactor track as
`v2`, so the team can evolve `lavoval` without losing working product momentum.

`lavoval` started as a marketplace-oriented foundation for exchanging human skills in the AI era.
`v2` extends that foundation into a platform with runtime execution, clearer package boundaries,
and a future CLI surface.

## V1 Baseline

`v1` is the current stable foundation and should be treated as the migration source, not as a
throwaway prototype.

### Product

- Public marketplace landing page
- Authentication: register, login, logout
- Account area
- Admin area
- Skill authoring and publishing flows
- Marketplace-oriented product copy

### Frontend

- `Next.js` app in `apps/web`
- Public, account, and admin routes
- Typed API client layer
- Reusable UI primitives
- Server actions for auth and mutations
- Route protection through `proxy.ts` and server-side guards

### Backend

- `Go` API in `apps/api`
- Layered architecture: handlers, services, repositories, middleware
- JWT-based auth
- RBAC with `user` and `admin`
- Public skills endpoints
- Authenticated "my skills" endpoints
- Admin users and admin skills endpoints

### Data

- `users`
- `profiles`
- `skills`
- `skill_modules`
- `user_skill_enrollments`

### Tooling

- `pnpm` workspace root
- Docker-backed local PostgreSQL
- SQL migrations and seeds
- Root scripts for dev, test, lint, and format
- Basic frontend and backend tests

## V2 Target

`v2` turns `lavoval` from a marketplace foundation into a platform plus runtime.

### Product direction

- Skills are not only listed and managed, but also executable
- Runtime becomes a first-class part of the product
- Users can run skills and view run history
- The platform grows toward engine, registry, adapters, and CLI

### Target architecture

```text
lavoval/
  apps/
    web/
    api/

  packages/
    engine/
    registry/
    adapters/
    sdk/

  infra/
    docker/
    db/
```

### Target naming

#### CLI

```bash
lavoval init
lavoval dev
lavoval run
lavoval deploy
lavoval skill create
lavoval skill run
```

#### Services

- `lavoval-engine`
- `lavoval-registry`
- `lavoval-runtime`

#### API direction

```text
/v1/auth
/v1/users
/v1/skills
/v1/modules
/v1/runtime/run
/v1/runtime/runs
```

## Migration Strategy

The main rule is: do not perform a cosmetic rewrite before runtime exists.

### Phase 1: Runtime inside the current monolith

Goal: add the missing product capability first.

- Keep the Go backend in a single app boundary while runtime is introduced
- Add runtime execution as new modules inside the current Go backend
- Add `skill_runs`
- Add mock executors
- Add runtime API
- Add runtime UI and run history

This gives real product value without destabilizing the repository.

### Phase 2: Structural refactor after runtime works

Goal: move from a working monolith to a cleaner platform layout.

- Move the backend filesystem layout to `apps/api`
- Rename infrastructure paths from `infrastructure` to `infra` only when scripts and docs are
  updated in the same pass
- Extract runtime logic into `packages/engine`
- Extract skill definition and management concerns into `packages/registry`
- Introduce `packages/adapters` only when adapter contracts are clear
- Introduce `packages/sdk` once API usage patterns stabilize

### Phase 3: CLI and platform edges

Goal: expose the platform through developer-facing commands.

- Add a real `lavoval` CLI
- Implement `lavoval dev`
- Implement `lavoval run <skill-id>`
- Add skill-related CLI subcommands
- Back the CLI with `sdk` rather than duplicating request logic

Current progress:

- initial CLI package exists
- `lavoval dev` is available
- `lavoval auth login`, `auth me`, and `auth logout` are available
- `lavoval skills list` is available
- `lavoval runs list` and `runs get` are available
- `lavoval admin runs list` is available
- `lavoval run <skill-id>` and `lavoval skill run <skill-id>` are available

## Design Principles For V2

- Keep `v1` running while `v2` is being introduced
- Prefer additive changes before filesystem moves
- Avoid abstracting adapters before two or more executors actually justify the split
- Keep runtime in-process for MVP-level `v2`
- Favor migration commits that are understandable and reversible
- Do not rename public concepts unless the product meaning changes

## What Changes In The Data Model

### Keep from v1

- `users`
- `profiles`
- `skills`
- `skill_modules`

### Add in v2

- `skill_runs`

### Extend `skills` gradually

The current `skills` table is useful and should be evolved, not replaced immediately.

Expected additions:

- `provider`
- `entrypoint`
- `config_json`
- optionally `updated_by`

This allows runtime execution without discarding the existing marketplace and authoring model.

## V1/V2 Compatibility Notes

- `skills` stay the main top-level product entity
- `modules` remain useful in `v2` as structured skill content
- existing auth and RBAC remain valid
- existing public/account/admin zones remain valid
- runtime becomes a new capability layered onto the current product

## First 10 V2 Tasks In Commit Order

These tasks are ordered to maximize product progress while minimizing churn.

### 1. `docs: define v2 runtime and refactor roadmap`

- Add and align documentation for `v1` baseline and `v2` target
- Freeze naming and migration rules
- Avoid architectural drift before implementation starts

### 2. `feat(db): add skill_runs table`

- Create migration for `skill_runs`
- Include statuses, input/output payloads, timestamps, and error storage
- Keep the schema compatible with in-process execution

### 3. `feat(skills): extend skill model for runtime metadata`

- Add `provider`
- Add `entrypoint`
- Add `config_json`
- Thread new fields through domain models, repositories, API DTOs, and contracts

### 4. `feat(runtime): add executor interface and registry`

- Introduce runtime package or module in the current backend
- Define executor contract
- Add registry lookup by `entrypoint`

### 5. `feat(runtime): implement mock executors`

- `echo`
- `text-summary-mock`
- `keyword-extract-mock`

These unlock the full runtime loop without external dependencies.

### 6. `feat(api): add runtime service and endpoints`

- `POST /api/v1/runtime/run`
- `GET /api/v1/runtime/runs`
- `GET /api/v1/runtime/runs/{id}`

Save run results in PostgreSQL.

### 7. `feat(web): add run skill flow from skill detail`

- Add a run form or action entry from the skill detail page
- Show run result in the UI
- Keep the experience simple and deterministic

### 8. `feat(web): add runs history pages`

- Add `/runs` or account-scoped runs UI
- Show status, timestamps, and output/error previews
- Allow users to inspect their own runs

### 9. `refactor(api): move backend into apps/api`

- Update scripts
- Update Dockerfiles
- Update docs
- Update imports and run commands

Do this only after runtime is already working.

### 10. `refactor(packages): extract engine, registry, and sdk`

- Extract stable runtime logic into `packages/engine`
- Extract skill management contracts into `packages/registry`
- Add `packages/sdk` for future CLI and external consumers

Only introduce `packages/adapters` if there is enough real adapter behavior to justify it.

## Deferred V2+ Work

These ideas are valid, but should not block the first `v2` milestone.

- full CLI surface
- deploy workflows
- external LLM providers
- queues and workers
- retries and orchestration
- teams and organizations
- billing
- advanced moderation
- versioned skill runtime graphs

## Definition Of Done For Early V2

Early `v2` is successful when:

- admin can create a runtime-capable skill
- a user can open a skill and run it
- the run is executed in-process through an executor registry
- the run is saved in `skill_runs`
- the result is visible in the UI
- the app still preserves the `v1` marketplace and admin foundation
