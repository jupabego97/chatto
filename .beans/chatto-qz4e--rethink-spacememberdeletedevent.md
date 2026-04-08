---
# chatto-qz4e
title: Rethink SpaceMemberDeletedEvent
status: todo
type: task
priority: critical
created_at: 2026-04-06T11:02:00Z
updated_at: 2026-04-06T11:02:00Z
parent: chatto-p1pf
---

SpaceMemberDeletedEvent naming is misleading (sounds like member kicked, but it's account deletion). Also question whether it should be space-scoped or instance-scoped. Options discussed: emit UserLeftSpace events per space, or consume instance-level UserDeletedEvent.
