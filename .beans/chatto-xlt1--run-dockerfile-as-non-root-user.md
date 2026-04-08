---
# chatto-xlt1
title: Run Dockerfile as non-root user
status: todo
type: task
priority: low
created_at: 2026-03-17T09:38:56Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzzw
parent: chatto-pmb4
---

Dockerfile.goreleaser has no USER directive — the process runs as root. Standard container hardening.

Fix: add adduser + USER directive.

**Files:** Dockerfile.goreleaser
**Severity:** Low
