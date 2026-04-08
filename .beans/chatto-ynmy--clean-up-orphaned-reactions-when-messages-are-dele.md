---
# chatto-ynmy
title: Clean up orphaned reactions when messages are deleted
status: draft
type: bug
created_at: 2026-01-24T16:40:54Z
updated_at: 2026-01-24T16:40:54Z
---

When a message is deleted, its reactions remain in the SPACE_{spaceId}_REACTIONS KV bucket. Over time this accumulates junk entries. DeleteMessage should cascade-delete all reactions for that message.