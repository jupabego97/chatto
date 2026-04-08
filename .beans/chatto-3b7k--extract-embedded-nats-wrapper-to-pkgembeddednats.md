---
# chatto-3b7k
title: Extract embedded NATS wrapper to pkg/embeddednats
status: todo
type: task
priority: normal
created_at: 2026-02-28T12:30:27Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzz
parent: chatto-w2dd
---

Extract the embedded NATS server wrapper from cli/internal/embedded_nats into cli/pkg/embeddednats.

## Scope

The package wraps nats-server with a Start() function that:
- Creates a NATS server from config options
- Starts it and waits for ready
- Integrates with errgroup for lifecycle management
- Provides InProcessConnectOption for TCP-free connections

## Decoupling needed

Currently imports config.EmbeddedNATSConfig. Replace with a self-contained options struct in the new package. The caller (main/cmd) maps from its config type to the pkg's options.

## Tasks
- [ ] Create cli/pkg/embeddednats package with its own config struct
- [ ] Move Start() and InProcessConnectOption
- [ ] Update cli/internal/embedded_nats/ callers to use new package
- [ ] Remove cli/internal/embedded_nats/
- [ ] Verify build and tests pass
