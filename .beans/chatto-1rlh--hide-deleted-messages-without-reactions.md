---
# chatto-1rlh
title: Hide deleted messages without reactions
status: draft
type: task
priority: normal
created_at: 2025-12-30T20:16:07Z
updated_at: 2026-01-02T20:46:15Z
parent: chatto-vjkr
---

Currently, deleted messages show a '[message deleted]' placeholder regardless of whether they have reactions. Messages that have been deleted AND have no reactions should be completely hidden (not rendered at all) instead of showing the placeholder.

## Rationale
- Deleted messages with reactions still provide value (the reactions are meaningful context)
- Deleted messages without reactions add visual noise without any value
- This matches Discord's behavior

## Implementation
- In the message rendering logic, check if message is deleted AND has no reactions
- If both conditions are true, don't render the message at all
- If deleted but has reactions, show the '[message deleted]' placeholder as we do now