---
# chatto-ec1m
title: Return DM permissions from server RBAC
status: todo
type: task
created_at: 2026-04-06T11:02:33Z
updated_at: 2026-04-06T11:02:33Z
parent: chatto-p1pf
---

Frontend hardcodes DM_PERMISSIONS object in Room.svelte, bypassing the server's viewerCan* fields for DM rooms. The server should return correct permissions for DM rooms so the frontend doesn't need special-casing.
