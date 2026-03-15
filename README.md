# Fitness Planning MVP

Privacy-first, mobile-first fitness planning web application with a Go backend, React frontend, Docker Compose runtime, PostgreSQL-backed state persistence, localized RU/KK UI, and deterministic fitness target calculations.

## What Is Implemented

- Public landing page with clear free-core-features message
- Email/password auth
- Google sign-in for login/registration with OAuth redirect flow
- Email verification
- Forgot/reset password
- Local-development verification/reset token exposure in `development` mode only
- Cookie-based sessions with CSRF protection for authenticated mutations
- Auth brute-force rate limiting
- Detailed onboarding form
- Deterministic calorie, macro, hydration, adherence, plan health, reschedule, and regeneration logic on the backend
- AI plan generation through an OpenAI-compatible adapter plus safe fallback provider
- Workout, meal, water, and weekly check-in tracking
- Plan versioning with history preservation in runtime state
- In-app notifications
- PostgreSQL relational projection over the declared schema with pgx-backed admin/trainer ops queries
- Trainer/admin/support/discussion MVP flows with deeper operational views
- RU and KK localization
- Light and dark themes with persisted device preference
- Docker Compose with `postgres`, `backend`, and `frontend`
- Separate compose entrypoints for `backend+postgres` and `frontend` standalone runs

## Current Architecture

- Frontend: React, TypeScript, Vite, React Router, TanStack Query, Zustand, SCSS
- Backend: Go, `net/http`, raw SQL migrations, structured logging, modular packages for app/domain/config/integrations
- Persistence:
  - SQL schema and seed catalog live in `backend/migrations`
  - Runtime app state is persisted to PostgreSQL through a raw-SQL snapshot store for MVP stability
  - The snapshot is projected into the relational schema so admin/trainer operational reads use pgx queries over real tables
- AI:
  - OpenAI-compatible HTTP adapter
  - strict privacy-safe input shaping in domain code
  - deterministic targets always stay on the backend

More detail is in:

- `docs/IMPLEMENTATION_PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DOMAIN_MODEL.md`
- `docs/API_CONTRACT.md`
- `docs/SECURITY_AND_PRIVACY.md`
- `docs/TEST_PLAN.md`

## Roles

- `user`
- `trainer`
- `admin`

## Folder Structure

- `frontend`
- `backend`
- `docs`
- root infra and env files

## Environment

Root:

- `.env.example`

Backend local:

- `backend/.env.example`

Frontend local:

- `frontend/.env.example`

E2E:

- root `package.json`
- `playwright.config.ts`
- `e2e/*.spec.ts`

Important variables:

- `APP_ADDR`
- `APP_PORT`
- `POSTGRES_HOST_PORT`
- `BACKEND_HOST_PORT`
- `FRONTEND_HOST_PORT`
- `DATABASE_URL`
- `SESSION_SECRET`
- `COOKIE_DOMAIN`
- `COOKIE_SECURE`
- `FRONTEND_URL`
- `BACKEND_URL`
- `CORS_ALLOWED_ORIGINS`
- `AI_API_BASE_URL`
- `AI_API_KEY`
- `AI_MODEL`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `ENABLE_NOTIFICATION_EMAILS`
- `HYDRATION_REMINDER_*`

For Gemini in OpenAI-compatible mode, a working example is:

- `AI_API_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai`

## Local Development

Backend:

```bash
cd backend
go test ./...
go test -race ./...
go vet ./...
go run ./cmd/api
```

Frontend:

```bash
cd frontend
npm install
npm test
npm run typecheck
npm run build
npm run dev
```

Playwright E2E:

```bash
npm install
npm run test:e2e:install
npm run test:e2e
```

Docker:

```bash
docker compose build
docker compose up -d
```

Default host ports in the root stack:

- frontend: `5173`
- backend: `8080`
- postgres: `5433`

Backend only with its own PostgreSQL:

```bash
cd backend
docker compose build
docker compose up -d
```

Frontend only:

```bash
cd frontend
docker compose build
docker compose up -d
```

Do not run the root stack and the split `backend/` or `frontend/` stacks at the same time unless you intentionally override ports.

## Verified Commands

These were executed successfully in the current environment:

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `npm test`
- `npm run typecheck`
- `npm run build`
- `docker compose build`
- `docker compose up -d`
- `docker compose ps`
- `docker compose -f backend/docker-compose.yml build`
- `docker compose -f frontend/docker-compose.yml build`

Smoke-tested flow:

- register
- verify email
- login
- onboarding
- generate plan
- restart backend
- login again and read persisted active plan
- Playwright user flow: register -> verify -> login -> onboarding -> generate plan -> water -> locale -> theme
- Playwright admin/trainer flow: admin assigns trainer and creates catalog records -> trainer reviews assigned user -> support reply creates notifications

## Assumptions

- v1 is a regular website, not a PWA
- browser push is intentionally not implemented in v1
- email delivery can be disabled during local development
- media is URL/metadata based
- a PostgreSQL snapshot-plus-relational-projection strategy is acceptable for MVP runtime reliability before a fully split repository-per-aggregate write model

## Security And Privacy Notes

- passwords are hashed with Argon2id
- session cookies are `HttpOnly`
- authenticated write requests require CSRF token matching
- auth endpoints use basic rate limiting
- verification and reset tokens are stored hashed
- privacy-safe AI payload shaping excludes direct identifiers such as email
- request IDs and security headers are applied centrally

Windows note:

- `gcc` was installed through `MSYS2`, and `C:\msys64\ucrt64\bin` was added to the user `PATH` so `go test -race ./...` works in new terminals as well

## Tradeoffs

- Writes still originate from the runtime application model and are projected into PostgreSQL; this is a pragmatic modular-monolith step, not a fully split command/query repository architecture.
- Trainer/admin/support/discussion UI is now materially deeper, but it is still an MVP operations console rather than a large internal enterprise tool.
- Email sending abstractions exist, but full provider wiring and log browsing remain intentionally lightweight.
- Some long-tail moderation and observability capabilities are still slimmer than the broadest possible interpretation of the original scope.

## Current Audit

Done:

- public landing, auth, verification, reset flow, onboarding, deterministic targets
- AI provider abstraction with strict JSON validation and privacy-safe payload shaping
- workout, meal, water, weekly check-in, notifications, plan versioning, plan health
- pgx-backed relational projection covering the declared schema surface used by the app
- SQL-backed admin users, trainers, support threads, discussion threads, notification logs, and trainer notes views
- RU/KK locale switching and light/dark themes
- root Docker Compose plus separate backend/frontend compose files
- Playwright E2E suite for key user/admin/trainer/support flows

Partially done:

- PostgreSQL write persistence still flows through snapshot + projection rather than isolated repository-per-aggregate commands
- trainer/admin/support/discussion capabilities are MVP-level rather than a full operational console
- AI output is validated strongly, but the persisted runtime plan model is still slimmer than the full original week/day/exercise schema

Not yet done:

- fully separated repository-per-aggregate write model across every bounded context
- fully expanded admin observability, moderation, and internal tooling promised in the largest interpretation of the original scope

## Development Seed Users

When `APP_ENV=development`, the backend auto-seeds these verified accounts for local QA and E2E:

- `admin@local.test`
- `trainer@local.test`
- `member@local.test`

Shared password:

- `DevPassw0rd!123`

## Future Improvements

- Replace snapshot persistence with table-level repositories over the existing schema
- Add richer AI generation request/response browsing, prompt/schema version tooling, and broader admin observability
- Expand trainer/admin moderation tools
- Add scheduler-backed reminder jobs instead of request-triggered reminder opportunities
