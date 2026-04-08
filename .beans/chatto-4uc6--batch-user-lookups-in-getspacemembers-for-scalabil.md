---
# chatto-4uc6
title: Batch user lookups in GetSpaceMembers for scalability
status: todo
type: task
priority: normal
created_at: 2026-01-28T09:29:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzz
---

## Problem

`GetSpaceMembers` in `spaces.go:1247` does O(N) individual `GetUser()` + `GetUserRoles()` calls per member. For spaces with many members, this generates a large number of sequential KV lookups with no batching.

## Solution

- For now, add a comment documenting the expected scaling limit
- Future: consider batch user lookups or a denormalized member list
- This is not urgent for small instances but should be addressed before supporting large spaces (1000+ members)

## Files

- `cli/internal/core/spaces.go` (GetSpaceMembers)
