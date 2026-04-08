---
# chatto-pwyl
title: Consolidate virtual role handling between HasSpaceUserPermissionViaRoles and PermissionResolver
status: todo
type: task
priority: normal
created_at: 2026-01-28T09:29:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzk
---

## Problem

`HasSpaceUserPermissionViaRoles` in `permissions.go:1006` manually checks the implicit `RoleEveryone` role by checking `SpaceMembershipExists` and then calling `engine.RoleHasPermission` for the everyone role separately.

Meanwhile, `PermissionResolver.getUserSpaceRoles` calls `c.core.GetUserRoles()` which already appends virtual roles (everyone + verified). These are two different implementations of the same concept, which could drift and cause subtle inconsistencies.

## Solution

- Refactor `HasSpaceUserPermissionViaRoles` to use `GetUserRoles()` (which includes virtuals) instead of manually handling the everyone role
- This keeps all virtual role resolution in one place

## Files

- `cli/internal/core/permissions.go` (HasSpaceUserPermissionViaRoles, HasSpaceUserPermissionDeniedViaRoles)
