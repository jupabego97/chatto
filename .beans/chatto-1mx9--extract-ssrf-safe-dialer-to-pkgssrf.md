---
# chatto-1mx9
title: Extract SSRF-safe dialer to pkg/ssrf
status: todo
type: task
priority: normal
created_at: 2026-02-28T12:30:17Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzk
parent: chatto-w2dd
---

Extract the SSRF-safe HTTP dialer from cli/internal/core/linkpreview/ssrf.go into cli/pkg/ssrf.

## Scope

The dialer blocks connections to private/reserved IP ranges (RFC 1918, loopback, link-local, etc.) by hooking into net.Dialer.Control. It has zero Chatto dependencies — only stdlib (net, syscall).

Useful for any service that fetches user-supplied URLs.

## Tasks
- [ ] Create cli/pkg/ssrf package
- [ ] Move SSRF dialer and IP range checks
- [ ] Move or adapt tests
- [ ] Update imports in linkpreview/
- [ ] Verify tests pass
