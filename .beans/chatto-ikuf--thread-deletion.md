---
# chatto-ikuf
title: Thread deletion
status: draft
type: feature
priority: normal
tags:
    - threading
    - backend
    - frontend
created_at: 2026-01-02T18:59:55Z
updated_at: 2026-01-02T20:45:43Z
parent: chatto-d2tk
---

Allow moderators to delete an entire thread and all its replies (cascade delete). Requires: new permission or reuse delete_any_message, deletion of all events with matching inThread, cleanup of ThreadMetadata KV entry, UI confirmation dialog.