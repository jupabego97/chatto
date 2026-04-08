---
# chatto-w2dd
title: Extract reusable packages to cli/pkg
status: in-progress
type: epic
priority: normal
created_at: 2026-02-28T12:29:33Z
updated_at: 2026-03-18T05:31:56Z
order: k
---

Refactor cli/internal to extract generic, non-business-logic code into a new cli/pkg directory. This improves reusability, enforces cleaner dependency boundaries, and reduces duplication (especially the KV cache pattern).

## Extraction Candidates

### Tier 1: Ready as-is
- pkg/kvstore — Generic JetStream KV bucket cache (highest value)
- pkg/spaserver — SvelteKit SPA serving middleware
- pkg/service — Service lifecycle interface
- pkg/encryption — ChaCha20-Poly1305 utilities
- pkg/graphql/depthlimit — gqlgen query depth limit extension
- pkg/signedurl — HMAC-signed transform URLs
- pkg/ssrf — SSRF-safe HTTP dialer

### Tier 2: Minor decoupling needed
- pkg/embeddednats — Embedded NATS server wrapper
- pkg/natsauth — NATS client auth options

### Tier 3: Needs design work
- pkg/rbac — Generic RBAC engine (protobuf coupling)
