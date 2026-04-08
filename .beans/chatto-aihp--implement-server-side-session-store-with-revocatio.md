---
# chatto-aihp
title: Implement server-side session store with revocation
status: todo
type: feature
priority: high
created_at: 2026-02-16T13:50:00Z
updated_at: 2026-03-18T05:34:10Z
order: "y"
parent: chatto-v29q
---

## Problem

Sessions are cookie-only with HMAC signing (gin-contrib/sessions cookie store). This means:

1. **No session revocation** — cannot invalidate sessions on password change or account compromise
2. **Signing secret leak = all sessions compromised** — no server-side validation
3. **No session enumeration** — cannot show users their active sessions
4. **90-day session lifetime** without re-auth for sensitive operations

## Relevant Files

- `cli/internal/http_server/server.go:88-98` — session store setup
- `cli/internal/http_server/auth.go` — login/logout handlers
- `cli/internal/http_server/graphql_auth.go` — session validation in GraphQL

## Approach

Replace cookie store with a NATS KV-backed session store:

1. Create a `SESSIONS` KV bucket with TTL matching session lifetime
2. Store session ID in cookie, session data in KV
3. On login: create KV entry with user_id, created_at, IP, user_agent
4. On logout: delete KV entry
5. On password change: delete all sessions for user
6. Validate session exists in KV on every request

This fits the existing architecture (NATS KV for everything) and supports the single-executable goal.

## Future Work

- "Active sessions" UI showing user's sessions
- Session metadata (IP, device) for anomaly detection
- "Remember me" vs short-lived sessions
