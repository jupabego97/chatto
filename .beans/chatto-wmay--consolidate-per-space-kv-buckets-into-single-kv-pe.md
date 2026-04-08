---
# chatto-wmay
title: Consolidate per-space KV buckets into single KV per space
status: draft
type: feature
priority: normal
created_at: 2026-03-19T10:21:17Z
updated_at: 2026-03-19T10:29:06Z
---

Reduce JetStream Raft group overhead by consolidating the 6 per-space KV buckets into a single `KV_SPACE_{id}` bucket with subject-prefixed keys.

## Motivation

Each R3 Raft group costs ~2-3 MB of RAM per cluster node. Currently each space creates 8 streams (6 KVs + 1 event stream + 1 object store). Consolidating the 6 KVs into 1 reduces per-space Raft groups from 8 to 3, a 62% reduction that becomes significant at scale (thousands of instances).

## Current → New

| Before | After |
|--------|-------|
| `KV_SPACE_{id}_BODIES` | `KV_SPACE_{id}` key: `bodies.{msgId}` |
| `KV_SPACE_{id}_CONFIG` | `KV_SPACE_{id}` key: `config.{key}` |
| `KV_SPACE_{id}_RBAC` | `KV_SPACE_{id}` key: `rbac.{key}` |
| `KV_SPACE_{id}_REACTIONS` | `KV_SPACE_{id}` key: `reactions.{key}` |
| `KV_SPACE_{id}_RUNTIME` | `KV_SPACE_{id}` key: `runtime.{key}` |
| `KV_SPACE_{id}_THREADS` | `KV_SPACE_{id}` key: `threads.{key}` |

Same treatment for DM space KVs.

## Migration

Offline migration via `chatto migrate` CLI command. Requires maintenance window (no running Chatto pods writing to old buckets).

Upgrade sequence:
1. Disable ingresses (no new traffic)
2. Scale down Chatto pods (no writers)
3. Run `chatto migrate` against live NATS cluster
4. Deploy new version (only knows consolidated KV)
5. Re-enable ingresses

Migration logic (idempotent, safe to re-run):
1. Connect to NATS, enumerate all accounts
2. For each account, read space list from KV_INSTANCE
3. For each space, check if old-style KV buckets exist
4. Iterate all keys, put into consolidated KV with prefixed key
5. Delete old bucket (frees Raft group)
6. Skip spaces already migrated

Data volumes are small (max_msgs_per_subject: 1, no history), migration should take seconds.

Future consideration: if zero-downtime migrations become necessary at scale, implement a dual-read strategy (write to new, read new-then-old fallback, cleanup after full rollout).

## Code changes needed

- [ ] Update KV bucket creation in space provisioning
- [ ] Update all KV access code to use prefixed keys
- [ ] Update KV watchers to use subject filters (e.g. `bodies.>`)
- [ ] Implement migrate-on-startup logic
- [ ] Update backup/restore if it references specific bucket names
- [ ] Update tests
- [ ] Also consolidate instance-level KVs where feasible (KV_USER_PRESENCE + KV_CALL_STATE into one memory-backed KV)

## Per-Key TTL

Enable `allow_msg_ttl: true` on the consolidated KV bucket. This allows setting TTL per individual `Put()` call, which is more flexible than the current per-bucket `max_age` approach.

- Set bucket `max_age: 0` (unlimited) on the consolidated KV
- Apply per-key TTLs on put: e.g. notification keys get 90-day TTL, auth tokens get 90-day TTL, link preview cache gets 48h TTL
- Keys without explicit TTL live indefinitely (bodies, config, rbac, etc.)
- Note: per-message TTL can only be shorter than the stream's `max_age`, not longer — so keep `max_age: 0` or set it to the longest needed TTL
