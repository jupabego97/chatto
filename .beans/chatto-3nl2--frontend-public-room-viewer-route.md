---
# chatto-3nl2
title: 'Frontend: public room viewer route'
status: todo
type: task
priority: normal
created_at: 2026-02-10T08:05:44Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzw
parent: chatto-2quz
blocked_by:
    - chatto-bwc7
---

Create a read-only frontend route for viewing public rooms without authentication.

## Implementation

### Route structure
Create a new route tree outside \`/chat/\` (which requires auth):

\`\`\`
frontend/src/routes/public/
  +layout.svelte          — minimal layout (no sidebar, no chat chrome)
  +layout.ts              — no auth required
  [spaceId]/[roomId]/
    +page.svelte          — public room viewer
    +page.ts              — extract params, load room data
\`\`\`

### Layout
- Minimal header with room name, space name, and a "Sign up to join the conversation" CTA
- No sidebar, no room list, no member list
- Chatto branding / link to instance root
- Responsive / mobile-friendly

### Room viewer component
- Render message history in the same visual style as the authenticated view
- Messages show: author display name, avatar, timestamp, body (with markdown), attachments
- Messages do NOT show: reactions (or show counts only), thread reply counts
- No message composer
- No context menus or action buttons on messages
- Scroll behavior: start at bottom (most recent), scroll up for history

### Polling for updates
- On mount, fetch room events via GraphQL query
- Set up a polling interval (e.g., every 30 seconds) to refetch
- Use a simple \`setInterval\` + \`graphqlClient.client.query()\` pattern
- Show a subtle "Last updated X seconds ago" indicator
- No WebSocket connection

### GraphQL queries
- Reuse existing \`room\` and \`roomEvents\` queries (backend handles auth bypass)
- The GraphQL client needs to work without auth cookies — verify this works with the current urql setup (it should, since the HTTP middleware already allows unauthenticated requests through)

### Error states
- Room not found → 404 page
- Room is private → "This room is private" message with login CTA
- Room is archived → "This room has been archived" message

## Key files (new)
- \`frontend/src/routes/public/+layout.svelte\`
- \`frontend/src/routes/public/+layout.ts\`
- \`frontend/src/routes/public/[spaceId]/[roomId]/+page.svelte\`
- \`frontend/src/routes/public/[spaceId]/[roomId]/+page.ts\`

## Key files (reference)
- \`frontend/src/routes/chat/[spaceId]/[roomId]/+page.svelte\` — existing room view to draw from
- \`frontend/src/lib/components/room/\` — existing message rendering components

## Tests
- E2E: unauthenticated user can view public room messages
- E2E: unauthenticated user sees "private" error for non-public rooms
- E2E: polling updates content after new message is posted
