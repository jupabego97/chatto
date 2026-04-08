---
# chatto-ia87
title: Reduce live event refetch waterfalls
status: todo
type: task
priority: high
created_at: 2026-04-06T11:02:33Z
updated_at: 2026-04-06T11:02:33Z
parent: chatto-p1pf
---

Reaction, edit, delete, and video processing live events don't carry updated data — the frontend must refetch each affected message individually. SpaceMemberDeleted triggers serial refetch of ALL visible messages. Include updated data in events where feasible to eliminate round-trips.
