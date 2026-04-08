---
# chatto-fyyr
title: Delete space functionality
status: todo
type: feature
priority: normal
created_at: 2026-01-18T22:27:30Z
updated_at: 2026-03-18T05:34:10Z
order: zzzy
---

## Overview

Allow users with the `space.delete` permission to delete a space entirely.

## Backend

- Add `deleteSpace` mutation to GraphQL schema
- Implement in Core API layer with proper authorization
- Clean up all associated data:
  - Rooms and messages
  - Memberships
  - Assets/attachments
  - Space stream and KV entries
- Consider soft delete vs hard delete

## Frontend

- Add delete button in space settings (visible only with `space.delete` permission)
- Confirmation modal with space name typed to confirm (destructive action)
- Redirect to home/space list after deletion
- Handle case where user is viewing the space when it's deleted

## Considerations

- This is a destructive, irreversible action - needs strong confirmation UX
- May want to notify space members before/during deletion
- Consider what happens to DM rooms if applicable
- Update any cached space lists after deletion
