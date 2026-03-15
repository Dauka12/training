# Architecture

## System Overview
The application is a modular monolith with clear boundaries:

- Transport: HTTP handlers, middleware, DTO validation, cookies, error envelopes.
- Application: use-case orchestration, transactions, policy enforcement.
- Domain: calculations, rescheduling, adherence, plan health, trigger evaluation, authorization rules.
- Persistence: raw SQL repositories and query services.
- Integrations: email sender, AI provider, clock, token generator, scheduler.

## Backend Modules
- `cmd/api`: process bootstrap.
- `internal/config`: environment loading and config validation.
- `internal/httpx`: API helpers, middleware, response helpers.
- `internal/auth`: users, sessions, tokens, password policies.
- `internal/profile`: onboarding and preferences.
- `internal/catalog`: exercises, equipment, translations, media metadata.
- `internal/plans`: deterministic targets, AI generation, versioning, scheduling.
- `internal/tracking`: workout, meal, water, weekly check-ins.
- `internal/notifications`: in-app notifications, preferences, reminders, email dispatch.
- `internal/support`: support threads and public discussions.
- `internal/admin`: admin actions, audit logs.
- `internal/trainer`: trainer dashboard and notes.
- `internal/db`: migrations, connection, transactions.

## Frontend Modules
- `src/app`: app shell, providers, router.
- `src/features/auth`
- `src/features/onboarding`
- `src/features/dashboard`
- `src/features/plans`
- `src/features/tracking`
- `src/features/notifications`
- `src/features/trainer`
- `src/features/admin`
- `src/features/support`
- `src/shared/ui`
- `src/shared/lib`
- `src/shared/i18n`
- `src/shared/theme`

## Security Boundaries
- Server is the source of truth for auth, roles, plan logic, and tracking logic.
- AI is treated as untrusted input.
- Tokens are generated server-side, hashed at rest, and compared in constant time.
- Cookies are httpOnly and SameSite-aware; CSRF tokens protect state-changing browser requests.

## Background Jobs
- Reminder scheduler runs in-process for MVP.
- Job records are not yet externalized to a queue.
- Design keeps notifier channels abstract for future web push.
