# Implementation Plan

## Scope
Build a production-minded MVP fitness planning web application as a modular monolith:
- `backend`: Go, PostgreSQL, raw SQL, migrations, tests-first.
- `frontend`: React, TypeScript, Vite, React Router, TanStack Query, Zustand, Zod, SCSS.
- Root infra: Docker Compose, env examples, docs.

## Assumptions
- v1 uses server-rendered email links with token-based verification and reset flows.
- Cookie-based session auth is used for browser clients.
- Notifications in v1 are in-app plus email; web push is deferred.
- AI generation uses an OpenAI-compatible API, but all deterministic targets are calculated locally.
- Trainer assignment is manual through admin UI.
- Public discussions can be disabled by config but the module exists in MVP.
- Media is metadata-only in v1 and references safe URLs.
- Reminder scheduling uses a simple backend polling scheduler suitable for one-instance MVP deployments.

## Phases
1. Documentation and acceptance criteria.
2. Backend and frontend test harness setup.
3. Core backend domain tests.
4. Auth tests.
5. Core backend implementation.
6. Auth implementation.
7. Onboarding and deterministic calculations.
8. Catalogs and seed data.
9. AI generation pipeline with strict schema validation.
10. Tracking, reminders, notifications.
11. Trainer/admin panels.
12. Support and discussions.
13. Docker hardening, docs, final verification.

## Delivery Strategy
- Prefer a smaller but complete vertical slice over breadth without reliability.
- Keep architecture explicit and boring.
- Test critical paths first and avoid speculative abstractions.
