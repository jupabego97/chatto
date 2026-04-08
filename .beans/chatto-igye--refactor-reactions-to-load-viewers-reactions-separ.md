---
# chatto-igye
title: Refactor reactions to load viewer's reactions separately
status: draft
type: task
priority: normal
created_at: 2025-12-28T14:05:09Z
updated_at: 2026-01-02T20:46:15Z
parent: chatto-vjkr
---

Currently reactions include user data for all reactions when loading. Refactor to:

1. **Remove user data from reaction counts** - When loading reactions for a message, only return the reaction type and count (e.g., `thumbsup: 5`), not the list of users who reacted

2. **Load viewer's reactions separately** - Add a separate field/query to get which reactions the current user has added, for UI state (highlighting the viewer's own reactions)

This reduces payload size and simplifies the data model while still allowing the UI to show:
- Total count per reaction type
- Which reactions the viewer has added (for toggle state)