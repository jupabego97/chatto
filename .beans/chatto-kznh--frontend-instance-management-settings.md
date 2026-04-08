---
# chatto-kznh
title: 'Frontend: Instance management settings'
status: draft
type: feature
priority: normal
created_at: 2026-03-03T11:49:09Z
updated_at: 2026-03-03T11:49:30Z
parent: chatto-wadw
blocked_by:
    - chatto-1uxi
    - chatto-qsk5
    - chatto-012a
---

Build UI for managing connected instances: viewing, editing, reordering, disconnecting, and re-authenticating.

## Context

Once users can add instances, they need a way to manage them. This includes seeing the list of connected instances, their status, and performing actions like disconnecting or re-authenticating.

## Requirements

- [ ] Instance management page (accessible from sidebar settings or a dedicated route like `/settings/instances`)
- [ ] List all registered instances with: name, URL, icon, connection status, authenticated user
- [ ] "Disconnect" / remove an instance (with confirmation)
- [ ] "Re-authenticate" — if a token expires or is revoked, allow re-login without removing the instance
- [ ] Edit instance display name / local slug
- [ ] Reorder instances (drag-and-drop or up/down buttons) — affects sidebar order
- [ ] Show connection health (last successful query time, WebSocket status)
- [ ] The "home" instance cannot be removed (it's the one serving the SPA) — but can be hidden or de-emphasized if desired
- [ ] Write e2e tests for instance management actions

## Implementation Notes

### Route
Add a settings page:
```
/settings/instances         — list and manage instances
/settings/instances/add     — add instance flow (could reuse the modal)
```

Or integrate into the existing `/chat/settings/` hierarchy. Since settings may become "global" (not per-instance), a top-level `/settings/` route makes more sense.

### Re-authentication flow
When a token expires:
1. The instance's GraphQL client detects an auth error
2. The instance is marked as "needs re-auth" in the registry
3. The sidebar shows a warning indicator on the instance icon
4. Clicking the indicator (or going to settings) opens the login flow for that instance

### Key files to create
- `frontend/src/routes/settings/instances/+page.svelte`
- Instance management components

### Blocked by
- Instance registry (need instance data)
- Add instance flow (for re-auth, reuse the auth portion)
- Multi-instance sidebar (shows connection status)
