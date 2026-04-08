---
# chatto-2quz
title: Public rooms (read-only, unauthenticated access)
status: draft
type: epic
created_at: 2026-02-10T08:05:01Z
updated_at: 2026-02-10T08:05:01Z
---

Allow space admins to mark individual rooms as **public** at creation time. Public rooms are viewable by anyone on the internet without an account — read-only, no interaction.

## Goals

- Unauthenticated users can view message history in public rooms
- Read-only: no posting, reactions, threads, or any write operations
- Periodic content refresh without WebSocket connections (polling first, SSE later)
- Privacy-safe: public/private is chosen at room creation time and cannot be changed later

## Key Design Decisions

### Encryption is NOT a blocker
Messages are encrypted at rest with the *author's* per-user key (ChaCha20-Poly1305) for GDPR crypto-shredding. The server already decrypts on behalf of authenticated readers. Public access uses the same server-side decryption path — no encryption model changes needed. If an author's key is crypto-shredded, their messages show as "[Message unavailable]" for everyone, including public readers. This is consistent and correct.

### Public/private is immutable after creation
To protect user privacy, a room's public/private status is set at creation time and cannot be changed. Users posting in a private room can trust it stays private. Admins can archive a public room, but can't retroactively make a private room public (or vice versa). This is the simplest approach and avoids complex "only show messages after the toggle" logic.

### Polling for live updates (v1)
Rather than adapting the WebSocket subscription infrastructure (heavily auth-coupled), v1 uses simple HTTP polling on the room events query. A dedicated SSE endpoint could be added later as an enhancement.

### Separate frontend route tree
Public room viewing lives outside `/chat/` (which requires authentication). A new `/public/[spaceId]/[roomId]` route renders a minimal, read-only view with no composer, reactions, or member list.

## Non-goals (for now)

- Real-time WebSocket subscriptions for public rooms
- Anonymous posting or interaction
- SEO optimization / server-side rendering
- Public spaces (only individual rooms)
- Shareable invite links (separate feature)
