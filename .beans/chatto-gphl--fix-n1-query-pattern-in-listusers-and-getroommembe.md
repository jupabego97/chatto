---
# chatto-gphl
title: Fix N+1 query pattern in ListUsers and GetRoomMembersList
status: todo
type: bug
priority: high
created_at: 2026-01-23T21:00:46Z
updated_at: 2026-03-18T05:34:10Z
order: w
---

### Problem

The `ListUsers` function in `cli/internal/core/users.go:435-458` and `GetRoomMembersList` in `cli/internal/core/rooms.go:869-905` exhibit N+1 query patterns that cause severe performance degradation at scale.

**ListUsers:** Fetches all user keys, then does a GET for EACH one:
```go
for key := range keyLister.Keys() {
    entry, err := c.storage.instanceKV.Get(ctx, key)  // N+1: Individual GET per user
}
```

With 1000 users, this makes **1001 KV calls** (1 list + 1000 gets).

**GetRoomMembersList:** Fetches ALL room memberships for the space, filters in memory:
- If a space has 100 users and 10 rooms, you fetch 1000 memberships to get ~10

### Impact
- Admin dashboard and member lists become progressively slower as user count grows
- Scalability blocker for production deployments

### Suggested Fix
- Use batch KV gets if NATS supports them
- Revise key patterns to enable server-side filtering via NATS wildcards
- Consider caching frequently-accessed counts

### Files to Modify
- `cli/internal/core/users.go` (ListUsers)
- `cli/internal/core/rooms.go` (GetRoomMembersList)

### Todo
- [ ] Research NATS batch get capabilities
- [ ] Implement optimized ListUsers
- [ ] Implement optimized GetRoomMembersList
- [ ] Add performance benchmarks
