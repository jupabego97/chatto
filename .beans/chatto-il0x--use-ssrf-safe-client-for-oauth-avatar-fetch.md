---
# chatto-il0x
title: Use SSRF-safe client for OAuth avatar fetch
status: todo
type: bug
priority: normal
created_at: 2026-03-17T09:38:45Z
updated_at: 2026-03-18T05:34:10Z
order: zzzV
parent: chatto-pmb4
---

fetchAndUploadUserAvatar uses http.DefaultClient with no SSRF protections, no timeout, no response size limit. The link preview fetcher correctly uses NewSSRFSafeClient for the same pattern — the right tool already exists in the codebase.

Fix: replace http.DefaultClient with linkpreview.NewSSRFSafeClient() and add io.LimitReader(resp.Body, 5*1024*1024).

**Files:** cli/internal/http_server/auth.go (lines 690-718)
**Severity:** Medium
