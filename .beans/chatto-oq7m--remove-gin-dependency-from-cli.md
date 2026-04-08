---
# chatto-oq7m
title: Remove Gin dependency from CLI
status: draft
type: task
priority: low
created_at: 2026-01-01T19:16:53Z
updated_at: 2026-01-02T20:46:15Z
parent: chatto-03l2
---

Evaluate replacing Gin with Go standard library (net/http) or a minimal router like chi.

## Context

Gin is used in 7 files with ~78 instances of Gin-specific APIs:
- `cli/cmd/start.go` - Sets Gin mode
- `cli/internal/http_server/server.go` - Core router setup
- `cli/internal/http_server/graphql.go` - GraphQL endpoint
- `cli/internal/http_server/auth.go` - Authentication routes
- `cli/internal/http_server/assets.go` - Asset serving
- `cli/internal/http_server/frontend.go` - Static frontend serving
- `cli/internal/http_server/graphql_auth.go` - Auth context injection

## Gin Features In Use

- **Routing**: `router.GET/POST/Any`, `router.Group()`, `:param` and `*wildcard` syntax
- **Handlers**: `gin.Context` for params, query, JSON binding, responses
- **Middleware**: `gin.Recovery()`, `gin.Logger()`, `gin-contrib/sessions`, `gin-contrib/static`
- **Server**: `gin.Engine` as `http.Handler`

## Migration Path

1. **Router**: Go 1.22+ `net/http` has basic path params, or use `chi` as middle ground
2. **Handlers**: Convert `func(c *gin.Context)` to `func(w http.ResponseWriter, r *http.Request)`
3. **Sessions**: Replace `gin-contrib/sessions` with `gorilla/sessions`
4. **Recovery**: Write custom panic recovery middleware
5. **Static files**: Use `http.FileServerFS()` with `embed.FS`
6. **Tests**: Rewrite all test helpers (significant effort)

## Considerations

- chi might be a better target than pure stdlib (stdlib-compatible signatures, easier migration)
- Significant test rewrite required
- Low priority - Gin works fine, this is mostly about reducing dependencies