---
# chatto-k6y2
title: Improve Boolean-returning mutation return types
status: todo
type: task
created_at: 2026-04-06T11:02:17Z
updated_at: 2026-04-06T11:02:17Z
parent: chatto-p1pf
---

joinSpace→Space, joinRoom→Room, RBAC mutations→Role. Only change where the server produces data the client doesn't already have. Leave Boolean for fire-and-forget operations (sendTypingIndicator, deleteMyAccount, etc.).
