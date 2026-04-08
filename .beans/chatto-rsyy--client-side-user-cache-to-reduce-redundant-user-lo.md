---
# chatto-rsyy
title: Client-side user cache to reduce redundant user lookups
status: draft
type: feature
priority: normal
created_at: 2025-12-30T21:33:50Z
updated_at: 2026-01-02T20:46:00Z
parent: chatto-vjkr
---

Introduce a client-side user store/cache that caches user information to reduce redundant GraphQL queries and server-side KV lookups.

## Problem

Currently, `roomToDMConversation` makes N separate `GetUser` calls for each participant. This is an N+1 query pattern that will cause performance issues at scale. Similar patterns exist elsewhere (room member lists, message author lookups, etc.).

## Proposed Solution

1. **Backend**: Add a `GetUsers(ctx, ids []string)` batch function that fetches multiple users in a single operation
2. **GraphQL**: Consider a DataLoader pattern for deferred batch loading
3. **Frontend**: Implement a client-side user cache that:
   - Stores user data by ID
   - Automatically populates from GraphQL responses
   - Provides `getUser(id)` that returns cached data or fetches if missing
   - Has reasonable TTL or invalidation strategy

## Benefits

- Reduces server load and latency for DM conversation lists
- Improves perceived performance for UI rendering
- Foundation for other features that need user data (mentions, presence, etc.)

## Related

This addresses the N+1 query issue identified in PR #48 code review.