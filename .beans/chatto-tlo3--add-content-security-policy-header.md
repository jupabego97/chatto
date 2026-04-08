---
# chatto-tlo3
title: Add Content-Security-Policy header
status: todo
type: task
priority: normal
created_at: 2026-03-17T09:38:52Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzw
parent: chatto-pmb4
---

Security headers are set for X-Content-Type-Options, X-Frame-Options, and Referrer-Policy, but no CSP. CSP is the most impactful single header for XSS mitigation, particularly important given {@html} usage in MessageContent.svelte.

Start with report-only mode to avoid breaking the SPA. Needs testing with SvelteKit to ensure inline styles work.

**Files:** cli/internal/http_server/frontend.go (lines 146-151)
**Severity:** Medium
