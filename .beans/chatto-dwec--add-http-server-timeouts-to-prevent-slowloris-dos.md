---
# chatto-dwec
title: Add HTTP server timeouts to prevent Slowloris DoS
status: todo
type: bug
priority: high
created_at: 2026-03-17T09:38:38Z
updated_at: 2026-03-18T05:34:10Z
order: k
parent: chatto-pmb4
---

http.Server instances lack ReadTimeout, WriteTimeout, ReadHeaderTimeout, and IdleTimeout. Vulnerable to Slowloris attacks where partial headers exhaust goroutines.

Quick fix: add ReadHeaderTimeout: 10s to all http.Server instances. WriteTimeout needs care for WebSocket/subscription routes.

**Files:** cli/internal/http_server/server.go (lines 156-166)
**Severity:** High
