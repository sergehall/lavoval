# Architecture Overview

Lavoval is organized as a growth-oriented fullstack workspace for a skill-exchange marketplace in the AI era, with explicit boundaries between presentation, application logic, data access, contracts, and infrastructure.

## Core boundaries

- `apps/web`: Next.js App Router frontend with public, authenticated account, and admin route trees.
- `backend/api`: Go REST API with handlers, services, repositories, auth, middleware, and PostgreSQL integration.
- `packages/contracts`: Typed TypeScript contracts for frontend-facing request and response shapes.
- `infrastructure`: Dockerfiles, database migrations, and seed data.
- `docs`: Product and technical onboarding documentation.

## Why this shape

- The frontend is isolated as a standalone app so it can evolve into a stronger marketplace, SSR, BFF, or multi-app delivery without coupling UI code to backend transport details.
- The backend keeps a layered structure: handlers know HTTP, services know use cases, repositories know persistence, and domain models stay framework-light.
- PostgreSQL schema supports soft delete, status-driven workflows, authorship, and learning relationships without hardcoding future product constraints.
- The project is intentionally prepared for future additions: matching, trust layers, organizations, permissions matrix, activity logs, billing, notifications, search, pagination, and background jobs.

## Auth model

- Backend issues JWT access and refresh tokens.
- Frontend server actions store tokens in HTTP-only cookies.
- `proxy.ts` protects account and admin routes early.
- Server layouts enforce the same rules on the server side.
- RBAC currently distinguishes `user` and `admin`, while the data model allows richer permission systems later.

## Product lens

- Skills are meant to be published by users, not only by administrators.
- Modules turn expertise into reusable learning units that can later support exchange, bundles, curation, and guided progression.
- Public and account areas separate discovery from participation, which fits a marketplace model.
- Admin capabilities exist to govern quality and trust as the exchange grows.

## Scale-oriented decisions

- Skill and module records are separate entities to support versioning, analytics, and editorial workflows.
- Enrollment/progress is modeled independently so assignments, cohorts, or org-based rollouts can be added later.
- Admin CRUD is isolated from public account flows to keep governance concerns explicit.
- Config comes from environment variables only; secrets are never hardcoded in application logic.
