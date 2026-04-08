---
# chatto-qa28
title: Evaluate SameSite=Strict for session cookies
status: todo
type: task
priority: low
created_at: 2026-02-16T13:52:58Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzz
parent: chatto-v29q
---

## Problem

Session cookies use `SameSite=Lax` mode (`cli/internal/http_server/server.go:96`). While Lax is a reasonable default, Strict provides stronger CSRF protection by preventing the cookie from being sent on any cross-site navigation.

## Considerations

- **Lax** allows cookies on top-level navigations (GET requests) from external sites — this means links to Chatto from other sites will include the session cookie
- **Strict** blocks cookies on all cross-site requests — users clicking a link to Chatto from an email or external site would need to re-authenticate
- Chatto doesn't appear to have flows that depend on cross-site cookie inclusion (no OAuth callbacks that need session cookies, no cross-origin form submissions)

## Recommended Approach

1. Evaluate whether any existing flows rely on Lax behavior
2. If no flows require it, switch to `SameSite=Strict` for maximum CSRF protection
3. Test login flows, email verification links, and any OAuth integrations

## Location

- `cli/internal/http_server/server.go` line 96: `SameSite: http.SameSiteLaxMode`

## Notes

- Low priority since Lax already provides good CSRF protection for state-changing requests
- The main trade-off is UX: users may need to log in again after clicking Chatto links from external sites
