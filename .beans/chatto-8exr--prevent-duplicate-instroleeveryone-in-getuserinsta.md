---
# chatto-8exr
title: Prevent duplicate InstRoleEveryone in getUserInstanceRoles
status: todo
type: task
priority: low
created_at: 2026-01-28T09:29:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzzk
---

## Problem

`PermissionResolver.getUserInstanceRoles` (permission_resolver.go:393) always appends `InstRoleEveryone` to the result of `GetUserInstanceRoles`. If the underlying method already includes 'everyone', this results in duplicate entries and redundant KV lookups during permission checks.

Not a correctness bug (key checks are idempotent), but a minor efficiency concern.

## Solution

- Check if 'everyone' is already in the list before appending, or
- Ensure `GetUserInstanceRoles` never includes it and document that contract

## Files

- `cli/internal/core/permission_resolver.go` (getUserInstanceRoles)
