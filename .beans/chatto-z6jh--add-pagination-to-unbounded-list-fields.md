---
# chatto-z6jh
title: Add pagination to unbounded list fields
status: todo
type: task
priority: high
created_at: 2026-04-06T11:02:17Z
updated_at: 2026-04-06T11:02:17Z
parent: chatto-p1pf
---

Room.members, Query.notifications, and Query.users return unbounded lists. Apply the SpaceMembersConnection pattern (limit/offset/totalCount/hasMore). Also consider threadEvents and Query.myFollowedThreads.
