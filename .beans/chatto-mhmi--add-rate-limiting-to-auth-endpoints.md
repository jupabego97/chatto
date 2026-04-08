---
# chatto-mhmi
title: Add rate limiting to auth endpoints
status: scrapped
type: task
priority: high
created_at: 2026-02-16T13:50:00Z
updated_at: 2026-04-08T16:02:12Z
order: z
parent: chatto-v29q
---

## Problem

The login (`/auth/login`), registration (`/auth/register`), password reset (`/auth/forgot-password`), and bootstrap (`/auth/bootstrap`) endpoints have no rate limiting. This exposes the application to:

- **Brute force attacks** on login
- **Credential stuffing** attacks
- **Account enumeration** via timing (despite dummy bcrypt comparison)
- **Password reset flood** (email spam)

## Relevant Files

- `cli/internal/http_server/auth.go` — all auth route handlers
- `cli/internal/http_server/server.go` — route setup

## Approach

Add a Gin middleware using a token bucket or sliding window rate limiter. Consider:

1. **Per-IP rate limiting** on all auth endpoints (e.g., 10 attempts/minute)
2. **Per-account rate limiting** on login (e.g., 5 attempts/minute per username)
3. **Global rate limiting** on registration (e.g., 20/hour per IP)
4. Return `429 Too Many Requests` with `Retry-After` header

Options: `golang.org/x/time/rate`, `github.com/ulule/limiter`, or a simple NATS KV-backed counter.

## References

- HTTP authn review finding: "No rate limiting on login, registration, or password reset endpoints"
- `cli/internal/http_server/auth.go`

## Reasons for Scrapping

Duplicate of rate limiting beans. Consolidated.
