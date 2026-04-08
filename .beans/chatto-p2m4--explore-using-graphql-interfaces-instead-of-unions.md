---
# chatto-p2m4
title: Explore using GraphQL interfaces instead of unions for NotificationItem
status: draft
type: task
created_at: 2026-01-21T21:28:09Z
updated_at: 2026-01-21T21:28:09Z
---

Evaluate whether NotificationItem should use a GraphQL interface instead of a union.

## Context

We currently use unions for polymorphic types (EventType, NotificationItem). This was chosen for simpler Go code (no getter scaffolding), but may not be the best choice for all cases.

## Interfaces vs Unions

**Interfaces are better when:**
- Types share meaningful common fields (like `id`, `createdAt`, `actor`)
- Clients frequently need those common fields together
- The shared fields represent a real abstraction

**Unions are better when:**
- Types are conceptually distinct with few/no shared fields
- You want simpler Go code (no getter scaffolding in gqlgen)

## NotificationItem Analysis

The notification types DO share common fields:
- `id` - notification ID
- `createdAt` - timestamp
- `actor` - user who triggered the notification
- `summary` - human-readable description

With an interface, client queries are cleaner:

```graphql
# Interface - cleaner query
query {
  notifications {
    id
    createdAt
    actor { id displayName }
    summary
    ... on MentionNotificationItem { space { id } }
    ... on ReplyNotificationItem { inReplyToId }
  }
}

# Union - more verbose (current)
query {
  notifications {
    ... on DMMessageNotificationItem { id createdAt actor { id } summary room { id } }
    ... on MentionNotificationItem { id createdAt actor { id } summary space { id } room { id } }
    ... on ReplyNotificationItem { id createdAt actor { id } summary space { id } room { id } inReplyToId }
  }
}
```

## Trade-offs

| Aspect | Interface | Union |
|--------|-----------|-------|
| Client DX | Better (shared fields at top level) | Worse (repeat fields in fragments) |
| Go code | More boilerplate (getter methods) | Simpler (marker methods only) |
| Type safety | Same | Same |
| Adding new types | Implement interface | Add to union |

## Recommendation

For `NotificationItem`: Interface is probably better since types share meaningful fields.
For `EventType`: Union is appropriate since event types are truly distinct.

## Tasks

- [ ] Evaluate if interface makes sense for NotificationItem
- [ ] If yes, refactor GraphQL schema to use interface
- [ ] Add getter methods to Go models
- [ ] Update frontend queries (simplify by removing repeated fields)
- [ ] Update rules/graphql.md with nuanced guidance (not just "prefer unions")