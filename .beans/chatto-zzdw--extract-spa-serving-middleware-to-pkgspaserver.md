---
# chatto-zzdw
title: Extract SPA serving middleware to pkg/spaserver
status: todo
type: task
priority: normal
created_at: 2026-02-28T12:29:53Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzV
parent: chatto-w2dd
---

Extract the generic SvelteKit SPA serving logic from cli/internal/http_server/frontend.go into cli/pkg/spaserver.

## What to extract

The reusable parts of setupFrontendRoutes():
- Precompressed file serving (.br, .gz variant detection + Content-Encoding headers)
- ETag generation from content-hash filenames (SvelteKit immutable assets)
- Immutable cache headers for /_app/immutable/ paths
- No-cache headers for HTML
- SPA fallback to 200.html for client-side routing
- Security headers (X-Frame-Options, X-Content-Type-Options, etc.)

## What stays in http_server

- OpenGraph meta tag injection (Chatto-specific, uses core to fetch space metadata)
- Session cookie refresh (Chatto-specific auth concern)

## Design

Expose as a Gin middleware/handler factory that accepts an embed.FS and optional config (fallback file name, immutable path prefix, etc.). The Chatto-specific logic wraps this with its OG injection and session refresh.

## Tasks
- [ ] Create cli/pkg/spaserver package
- [ ] Implement generic SPA handler with precompressed file support
- [ ] Add unit tests (immutable caching, precompressed serving, SPA fallback)
- [ ] Refactor frontend.go to use pkg/spaserver, keeping OG and session logic local
- [ ] Run e2e tests to verify no regressions
