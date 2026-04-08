---
# chatto-bfeb
title: Add pagination to spaces and users GraphQL queries
status: todo
type: task
priority: normal
created_at: 2026-02-16T13:52:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzV
parent: chatto-v29q
---

## Problem

The `spaces` and `users` queries return unbounded result sets. As the instance grows, these queries become increasingly expensive and could cause memory/performance issues.

## Locations

- `cli/internal/graph/schema.resolvers.go` — `Spaces()` and `Users()` resolvers
- `cli/internal/graph/schema.graphqls` — Query type definitions

## Recommended Approach

1. Add `first`/`after` cursor-based pagination arguments to both queries in the GraphQL schema
2. Implement cursor-based pagination in the resolvers using JetStream KV list operations
3. Set a reasonable default page size (e.g., 50) and maximum (e.g., 200)
4. Return a connection type with `edges`, `pageInfo` (hasNextPage, endCursor)

## Notes

- This is a GraphQL schema change (potentially breaking for clients), but acceptable at this stage per project rules
- The `users` query is admin-only, so impact is limited, but `spaces` is public for discovery
