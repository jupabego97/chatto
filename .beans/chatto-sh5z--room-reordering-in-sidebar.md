---
# chatto-sh5z
title: Room reordering in sidebar
status: todo
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2025-12-07T20:54:28Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzw
parent: chatto-12xq
---

Allow users to reorder rooms in the sidebar.\n\nOptions to consider:\n- Option A: Add position field to RoomMembership (simple, single source of truth)\n- Option B: Store ordered array of room IDs in separate KV bucket (atomic reorders, but sync issues)
