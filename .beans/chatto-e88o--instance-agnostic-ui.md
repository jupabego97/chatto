---
# chatto-e88o
title: Instance-Agnostic UI
status: in-progress
type: epic
priority: normal
created_at: 2026-03-12T16:36:33Z
updated_at: 2026-03-18T05:31:56Z
order: w
---

Treat all connected instances equally in the UI. The SPA is served by one instance (origin), but that's a transport detail, not a UI concept. Five phases: eliminate singletons, per-instance identity, hostname-scoped routing, feature parity, remove isHome. Supersedes the approach in chatto-wadw.
