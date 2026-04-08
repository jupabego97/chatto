---
# chatto-l573
title: Add rate limiting to authentication endpoints
status: scrapped
type: task
priority: high
created_at: 2026-03-17T09:38:35Z
updated_at: 2026-04-08T16:02:12Z
order: zV
parent: chatto-pmb4
---

No rate limiting exists anywhere in the codebase. /auth/login, /auth/register, /auth/forgot-password, /auth/reset-password, /auth/bootstrap are all unprotected. The timing-attack protection (dummy bcrypt for non-existent users) means every login attempt incurs bcrypt cost, amplifying DoS potential.

Consider per-IP rate limiting using golang.org/x/time/rate or ulule/limiter middleware.

**Files:** cli/internal/http_server/auth.go (lines 211-280, 286-349, 607-646)
**Severity:** High

## Reasons for Scrapping

Duplicate of rate limiting beans. Consolidated.
