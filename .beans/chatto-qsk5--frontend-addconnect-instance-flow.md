---
# chatto-qsk5
title: 'Frontend: Add/connect instance flow'
status: in-progress
type: feature
priority: normal
created_at: 2026-03-03T11:47:58Z
updated_at: 2026-03-18T05:31:56Z
order: zk
parent: chatto-wadw
blocked_by:
    - chatto-mxvw
    - chatto-qx5a
    - chatto-06x1
    - chatto-1uxi
---

Build the UI for adding a new Chatto instance to the client. This is the user-facing "onboarding" for connecting to additional instances.

## Context

Users need a way to add instances beyond the home instance. The flow: enter URL → probe instance → authenticate → add to registry → show in sidebar.

## Requirements

- [ ] "Add Instance" button in the sidebar opens the add-instance flow
- [ ] Step 1: Enter instance URL (text input with URL validation)
- [ ] Step 2: Probe the instance (fetch instance info endpoint — name, icon, auth methods)
- [ ] Show instance name and icon for confirmation
- [ ] Handle probe failures gracefully (instance unreachable, not a Chatto instance, CORS blocked)
- [ ] Step 3: Authenticate on the instance
  - If password auth available: show login form (email/password)
  - If OAuth available: show OAuth buttons (opens popup/redirect)
  - If registration available: show registration option
- [ ] Step 4: On successful auth, receive token, add instance to registry
- [ ] Auto-fetch space list after adding
- [ ] Instance appears in sidebar immediately
- [ ] Handle edge cases: instance already added, auth failure, network errors
- [ ] Write e2e tests for the add-instance flow

## UI Design

Could be implemented as:
- A modal/dialog (simpler, consistent with other modals in the app)
- A dedicated page (e.g., `/instances/add`) — works better for OAuth redirects

**Recommendation: Modal with a fallback to popup for OAuth.** The modal handles URL entry and password login. For OAuth, open a popup window to the instance's OAuth endpoint, which redirects back with a token.

## Implementation Notes

### URL normalization
- Strip trailing slashes
- Add `https://` if no protocol specified
- Validate it's a valid URL

### Instance probing
```typescript
const response = await fetch(`${normalizedUrl}/api/instance`);
const info = await response.json();
// Display name, icon, available auth methods
```

### Cross-origin login
For password auth, POST to `${instanceUrl}/auth/login` with credentials. The response includes the token (from the token-based auth bean).

For OAuth, this is trickier:
- Open `${instanceUrl}/auth/google?redirect=${ourOrigin}/auth/callback` in a popup
- The popup redirects back with `?token=...`
- The parent window reads the token from the popup's URL or via `postMessage`

### Key files to create
- `frontend/src/lib/components/AddInstanceModal.svelte` — the flow UI
- Supporting components for each step

### Blocked by
- Backend token-based auth (need tokens from login)
- Backend instance info endpoint (need to probe instances)
- Backend CORS support (need cross-origin fetch)
- Instance registry (need to store the result)

### Does NOT include
- Instance management (edit, remove) — separate bean
- Instance reordering — separate bean or part of sidebar bean
