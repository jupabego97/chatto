---
# chatto-qp50
title: Move user preferences from localStorage to KV storage
status: draft
type: task
created_at: 2026-01-21T21:21:52Z
updated_at: 2026-01-21T21:21:52Z
---

Currently user preferences (like `muteNotificationSounds`) are stored in localStorage, which means they don't sync across devices.

## Current State
- `frontend/src/lib/state/userPreferences.svelte.ts` stores preferences in localStorage
- Preferences are device-specific, not user-specific

## Proposed Changes
1. Create a KV bucket for user preferences (e.g., `USER_PREFERENCES`)
2. Add GraphQL queries/mutations for user preferences
3. Update frontend to read/write via GraphQL instead of localStorage
4. Consider fallback to localStorage for unauthenticated state

## Benefits
- Preferences sync across all user devices
- Consistent with other user data storage patterns
- Enables future preference types (notification preferences per space/room, etc.)

## Related
- Notification preferences feature (`chatto-2a9q`) may build on this foundation