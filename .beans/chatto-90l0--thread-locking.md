---
# chatto-90l0
title: Thread locking
status: draft
type: feature
priority: normal
tags:
    - threading
    - backend
    - frontend
created_at: 2026-01-02T18:59:54Z
updated_at: 2026-01-02T20:45:44Z
parent: chatto-d2tk
---

Allow moderators to lock a thread (prevent new replies). Requires: new permission (lock_thread), locked state in ThreadMetadata or message event, UI indicator for locked threads, disabled reply input when locked.