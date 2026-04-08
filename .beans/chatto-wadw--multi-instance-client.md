---
# chatto-wadw
title: Multi-Instance Client
status: draft
type: epic
created_at: 2026-03-03T11:45:01Z
updated_at: 2026-03-03T11:45:01Z
---

Transform the Chatto web UI from a single-instance SPA (tightly coupled to the backend that serves it) into a multi-instance client that can connect to multiple Chatto instances simultaneously.

## Vision

The left sidebar shows instance icons and their space icons together. Users can add instances by URL, authenticate independently on each, and see all their spaces across all instances in one unified view. The client can be:
- Embedded in a Chatto backend (current model, auto-connects to host instance)
- Hosted on a static CDN
- Shipped in a desktop app (e.g., Tauri)

## Current State

The frontend is deeply coupled to a single backend:
- GraphQL URL is a hardcoded relative path (`/api/graphql`)
- Auth is cookie-based (HttpOnly, SameSite=Lax) — cannot work cross-origin
- 13 module-level singleton stores assume one instance
- No CORS middleware on the backend
- Asset URLs are relative paths
- localStorage keys are not instance-scoped
- Routes have no instance segment

## Architecture Decisions

### Token-based auth (parallel to cookies)
Add bearer token auth as an alternative to session cookies. Cookies remain the default for the embedded SPA case (zero-config). Tokens are used for cross-origin connections. Tokens are stored per-instance in localStorage by the client.

### Instance registry (client-side)
A client-side data structure stored in localStorage. Each entry: `{ url, slug, name, iconUrl, token, userId }`. The embedded instance is auto-registered on first load.

### Active instance model
All instances maintain lightweight event subscriptions for unread counts and notifications. The "active" instance (the one being viewed) gets full message/presence subscriptions. This avoids N full WebSocket connections.

### Backward compatibility
The embedded SPA case should work identically to today with zero configuration. Multi-instance features activate when a user adds a second instance.

## Sidebar Design

```
[Instance 1 icon]
  [DM]
  [Space A icon]
  [Space B icon]
[Instance 2 icon]
  [DM]
  [Space C icon]
[+ Add Instance]
[Settings gear]
[User avatar]  ← local/UI settings (not instance-specific)
```

## Child Beans (in dependency order)

See child beans for detailed implementation plans.
