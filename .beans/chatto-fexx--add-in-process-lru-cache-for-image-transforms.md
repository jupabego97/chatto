---
# chatto-fexx
title: Add server-side caching for image transforms
status: draft
type: task
priority: low
created_at: 2025-12-15T16:52:12Z
updated_at: 2026-01-02T20:46:14Z
parent: chatto-vjkr
---

The dynamic image transformation endpoint (`/assets/space/{spaceId}/attachments/{attachmentId}/t/{signedPath}`) currently has no server-side caching - it relies entirely on CDN/proxy caching per ARCHITECTURE.md.

**Problem:** If running without a CDN (e.g., during development or small deployments), repeated requests for the same transformed image will re-read from object store and re-transform every time. This is wasteful for CPU and I/O.

## Options

### Option A: JetStream In-Memory KV (preferred)

Use a NATS JetStream KV bucket with memory storage:
- Key: `{spaceId}.{attachmentId}.{base64params}`
- Value: transformed image bytes
- Shared across all instances automatically
- Built-in TTL support for eviction
- Already using NATS, no new dependencies

Tradeoff: network hop to NATS vs in-process, but still much faster than re-reading from object store and re-encoding.

### Option B: In-process LRU cache

Use something like `hashicorp/golang-lru`:
- Purely in-memory, no network hop
- Each instance warms independently (no sharing)
- Need to manage memory budget explicitly

## Considerations

- **Optional and configurable**: Some deployments will have a CDN or proxy cache (nginx, Varnish, Cloudflare) in front of the app, making server-side caching redundant. This should be configurable via `[core.assets]` in the config file:
  - `transform_cache = "none"` (default) - no server-side caching, rely on external cache
  - `transform_cache = "memory"` - JetStream in-memory KV
  - `transform_cache = "lru"` - in-process LRU (if we implement it)
- Cache key: `{spaceId}/{attachmentId}/{base64params}` (or dotted for KV)
- Cache invalidation: attachments are immutable (same ID = same content), so no invalidation needed
- Memory pressure: need sensible defaults (TTL-based eviction for KV, size-based for LRU)