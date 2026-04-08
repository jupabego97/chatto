---
# chatto-ieqy
title: Add max password length check
status: todo
type: bug
priority: low
created_at: 2026-03-17T09:38:54Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzV
parent: chatto-pmb4
---

No max password length enforced. bcrypt silently truncates at 72 bytes, so passwords sharing the same first 72 bytes hash identically.

Fix: add len(password) > 128 check in both CreateUser and ResetPassword paths.

**Files:** cli/internal/core/users.go (line 51), cli/internal/http_server/auth.go (line 663)
**Severity:** Low
