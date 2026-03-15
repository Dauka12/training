# API Contract

## Base
- Prefix: `/api/v1`
- Content type: `application/json`
- Error envelope:

```json
{
  "error": {
    "code": "validation_error",
    "message": "Validation failed",
    "fields": {
      "email": "Invalid email"
    },
    "request_id": "req_123"
  }
}
```

## Public Endpoints
- `GET /healthz`
- `GET /readyz`
- `POST /auth/register`
- `POST /auth/verify-email`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`

## Authenticated Endpoints
- `GET /me`
- `PUT /me/preferences`
- `PUT /onboarding`
- `POST /plans/generate`
- `GET /dashboard/today`
- `GET /plans/active`
- `POST /tracking/workouts/:scheduleID/log`
- `POST /tracking/meals`
- `POST /tracking/water`
- `GET /tracking/hydration/summary`
- `POST /checkins/weekly`
- `GET /notifications`
- `POST /notifications/:id/read`

## Trainer Endpoints
- `GET /trainer/users`
- `GET /trainer/users/:id`
- `POST /trainer/users/:id/notes`
- `POST /trainer/users/:id/regenerate-plan`

## Admin Endpoints
- `GET /admin/users`
- `POST /admin/trainers/assign`
- `GET /admin/catalog/equipment`
- `POST /admin/catalog/equipment`
- `GET /admin/catalog/exercises`
- `POST /admin/catalog/exercises`
- `GET /admin/ai/logs`
- `GET /admin/email/logs`
- `GET /admin/notification/logs`
- `GET /admin/audit-logs`

## Support Endpoints
- `GET /support/threads`
- `POST /support/threads`
- `POST /support/threads/:id/messages`
- `GET /discussions/threads`
- `POST /discussions/threads`
- `POST /discussions/threads/:id/replies`
