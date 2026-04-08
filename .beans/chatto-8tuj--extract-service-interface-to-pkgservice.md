---
# chatto-8tuj
title: Extract service interface to pkg/service
status: todo
type: task
priority: low
created_at: 2026-02-28T12:29:57Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzzV
parent: chatto-w2dd
---

Move the Service interface from cli/internal/service to cli/pkg/service.

## Scope

This is a trivial move. The package contains a single interface:
```go
type Service interface {
    Run(ctx context.Context) error
}
```

Zero dependencies. Just move the file and update imports.

## Tasks
- [ ] Move cli/internal/service/ to cli/pkg/service/
- [ ] Update all import paths (http_server, video, etc.)
- [ ] Remove cli/internal/service/
- [ ] Verify build succeeds
