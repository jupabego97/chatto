---
# chatto-1agt
title: 'Security hardening: lower-priority items'
status: todo
type: task
priority: low
created_at: 2026-03-17T09:39:03Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzzy
parent: chatto-pmb4
---

Collected lower-priority findings from the March 2026 security audit. Each is individually small.

- [ ] Add HSTS header when configured URL uses HTTPS (cli/internal/http_server/frontend.go)
- [ ] Use URL fragment instead of query parameter for OAuth token redirect (cli/internal/http_server/auth.go:146-152)
- [ ] Gate GraphQL introspection/playground behind config flag (cli/internal/http_server/graphql.go:186,217-220)
- [ ] Require auth on token revocation endpoint or restrict to self-revocation (cli/internal/http_server/auth.go:190-207)
- [ ] Return generic error from /readyz instead of internal details (cli/internal/http_server/health.go:30-35)
- [ ] Validate client-provided link preview fields (enum, length limits) (cli/internal/graph/mutation.resolvers.go:341-364)
- [ ] Add AAD to AEAD encryption (breaking change, needs migration) (cli/internal/encryption/encryption.go:45,68)
- [ ] Add size bound to OpenGraph metadata cache (cli/internal/http_server/opengraph.go:38-62)
- [ ] Document reverse proxy must set/strip X-Forwarded-Host (cli/internal/http_server/graphql.go:97-98)
