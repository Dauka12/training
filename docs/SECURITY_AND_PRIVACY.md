# Security And Privacy

## Principles
- Privacy by design.
- Data minimization for AI payloads.
- Secure defaults for cookies, tokens, rate limits, and logging.
- Explicit authorization checks on every sensitive route.

## MVP Controls
- Argon2id password hashing.
- Hashed verification/reset tokens.
- Session rotation on login and privilege-sensitive events.
- CSRF protection for cookie-authenticated state changes.
- Request ID middleware.
- Security headers.
- Safe CORS allowlist.
- Brute-force protection on auth routes.
- Audit logs for admin-sensitive actions.
- Structured logs with redaction.

## Threat Model Notes
- Credential stuffing against login and reset flows.
- Session theft via XSS or weak cookies.
- IDOR on trainer/admin endpoints.
- Prompt injection or malformed AI responses.
- Leakage of PII in logs or AI payloads.
- Unauthorized access to historical plans and support threads.

## Deferred Improvements
- Encryption at rest for selected fields.
- External job queue.
- Multi-instance distributed scheduler lock.
- Web push channel.
