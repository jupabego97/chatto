---
# chatto-52e4
title: Clear subscription memberRooms on space kick
status: todo
type: bug
priority: high
created_at: 2026-03-17T09:38:41Z
updated_at: 2026-03-18T05:34:10Z
order: s
parent: chatto-pmb4
---

When a user is kicked from a space (SpaceMemberDeleted event), only presenceMemberCache is cleared. The memberRooms map is not touched, so the kicked user's subscription continues receiving room events until they disconnect.

Fix: add memberRooms clearing when memberDeleted.UserId == userID in the SpaceMemberDeleted handler (~4 lines).

**Files:** cli/internal/core/core.go (lines 1474-1476)
**Severity:** Medium
