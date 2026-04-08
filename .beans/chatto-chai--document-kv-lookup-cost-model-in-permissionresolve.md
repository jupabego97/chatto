---
# chatto-chai
title: Document KV lookup cost model in PermissionResolver
status: todo
type: task
priority: low
created_at: 2026-01-28T09:29:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzy
---

## Problem

`HasSpacePermission` and `HasRoomPermission` in permission_resolver.go perform up to ~16 individual KV Get calls per permission check (deny checks for all role lists, then grant checks in authority order). This is fine for in-process NATS KV but would be problematic if KV access ever becomes remote.

## Solution

- Add a doc comment to the PermissionResolver explaining the cost model
- Note that this is designed for in-process NATS KV latency characteristics
- If remote KV is ever needed, batch lookups or caching would be required

## Files

- `cli/internal/core/permission_resolver.go`
