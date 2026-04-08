---
# chatto-noia
title: Optimize reaction ordering with cached projection
status: draft
type: task
priority: normal
created_at: 2025-12-14T20:41:54Z
updated_at: 2026-01-02T20:46:15Z
parent: chatto-vjkr
---

Currently `GetReactions` fetches each reaction entry individually to read the timestamp for ordering. This works but could be slow for messages with many reactions.

## Potential Optimization

Cache a projection of reaction ordering in a memory-only KV bucket (or similar). For example:
- Key: `msg.{seqId}.order`
- Value: Ordered list of emoji names by first-added time

This projection would be updated when reactions are added/removed, avoiding the need to fetch all entries on every read.

## Considerations

- Memory KV bucket would be fast but ephemeral (lost on restart)
- Could rebuild projection lazily on first access after restart
- May not be worth the complexity unless reaction counts per message get large
- Current implementation is simple and correct - optimize only if needed

## Current Implementation

See `cli/internal/core/reactions.go` - `GetReactions` function fetches each entry to read timestamps stored as 8-byte big-endian uint64 nanoseconds.