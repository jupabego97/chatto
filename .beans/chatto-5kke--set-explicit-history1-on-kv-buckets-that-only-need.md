---
# chatto-5kke
title: Set explicit History:1 on KV buckets that only need current value
status: todo
type: task
priority: low
created_at: 2026-02-16T13:52:58Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzzz
parent: chatto-v29q
---

## Problem

Some KV buckets that only store current state (e.g., presence, read markers) don't set an explicit `History` value in their bucket config. NATS KV defaults to `History: 1`, but being explicit is better for documentation and prevents accidental history retention if defaults ever change.

## Recommended Approach

1. Audit all `CreateKeyValue` / `CreateOrUpdateKeyValue` calls
2. For buckets that only need current value (presence, typing indicators, read state), set `History: 1` explicitly
3. For buckets that benefit from history (e.g., permission audit trail), document why higher history is set

## Notes

- This is a minor cleanup — NATS already defaults to History:1
- Mostly a documentation/intent-clarity improvement
