# ADR-029: Destination Schema for the Server Model — Buckets, Streams, Subjects

**Date:** 2026-05-04

## Context

[ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) collapses Chatto's two-layer Instance + Space model into a single Server. That decision is largely about how operators and users *think* about the system, but it has a concrete, irreversible consequence at the storage layer: today's per-space JetStream resources (`SPACE_{id}_EVENTS` stream and the `SPACE_{id}_*` bucket family per [ADR-013](ADR-013-per-space-stream-sharding.md)) and the side-by-side instance buckets (`INSTANCE`, `INSTANCE_ASSETS`, `INSTANCE_RBAC`, `INSTANCE_CONFIG`, plus several runtime/cache/notification buckets) all need to be replaced by a single, server-wide layout.

The old layout exists for good reasons: per-space sharding isolates noisy spaces from quiet ones, separate body / reaction / thread buckets isolate hot-write workloads from structural state, separate cache buckets keep regeneratable data out of backups. None of those reasons disappear with the consolidation — only the per-space-ness does. The new layout therefore keeps the *axes* of the old layout (structural vs. body vs. reaction vs. thread vs. runtime vs. encryption-keys vs. presence) but flattens the per-space dimension out of every name and key.

This ADR locks the destination shape: which buckets exist, what's in each, what the JetStream streams look like, what subject patterns events use, and what the object store looks like. Phase 2 builds against this shape; Phase 4's migration tool transforms old data into this shape.

## Decision

### KV buckets — seven total

| Bucket | Backed up | Wipeable | Storage | Contents |
|---|---|---|---|---|
| `SERVER` | yes | no | File | Structural state: users, RBAC, rooms, room memberships, server config, branding metadata. |
| `RUNTIME` | **no** | **yes** | File, per-key TTL | Caches and transactional tokens: auth tokens, notifications, link previews, asset cache, read state, mention status, email-verification / password-reset / account-deletion tokens. |
| `BODIES` | yes | no | File | Message body content. Compound key `{userId}.{eventId}` so crypto-shredding a user invalidates every body in one key delete (per [ADR-007](ADR-007-per-user-encryption-with-crypto-shredding.md)). |
| `REACTIONS` | yes | no | File | Reaction aggregates per `{roomId}.{eventId}.{emoji}`. Separate bucket because reaction write volume swamps structural state. |
| `THREADS` | yes | no | File | Thread metadata per `{roomId}.{rootEventId}`. Separate bucket because the access pattern is read-modify-write under optimistic locking ([ADR-016](ADR-016-occ-for-message-publishing.md)), distinct from message read-many semantics. |
| `ENCRYPTION_KEYS` | **no (security)** | no | File | Per-user encryption keys. Explicitly excluded from `chatto backup` so that backups never contain the means to decrypt the data they hold. |
| `PRESENCE` | n/a | n/a | Memory + 60s TTL | Presence status only. Memory-backed with explicit `MaxBytes` to bound RAM. Wipe is automatic on restart; no backup story needed. |

Total: seven buckets, replacing 11+ instance-level + 8 per-space buckets in the old layout.

### Stream — one

| Stream | Subjects | Notes |
|---|---|---|
| `CHAT_EVENTS` | `room.>`, `joined`, `left`, `member_deleted` | One server-wide stream. Replaces all per-space `SPACE_{id}_EVENTS`. Subject filter is explicit, not `>`, so the stream can't accidentally swallow live or audit subjects. |

`AUDIT_EVENTS` is reserved for a future audit log stream and is **not** implemented as part of this consolidation.

### Object store — one

| Object store | Key prefixes | Replaces |
|---|---|---|
| `ASSETS` | `avatar/{userId}`, `branding/logo`, `branding/banner`, `attachment/{id}`, `attachment/{id}/thumb` | `INSTANCE_ASSETS` + every `SPACE_{id}_ASSETS`. |

### Subject patterns

All room and structural subjects drop the `space.{spaceId}.` prefix:

| Old | New |
|---|---|
| `space.{spaceId}.room.{roomId}.msg.{eventId}` | `room.{roomId}.msg.{eventId}` |
| `space.{spaceId}.room.{roomId}.msg.{rootId}.replies.{eventId}` | `room.{roomId}.msg.{rootId}.replies.{eventId}` |
| `space.{spaceId}.room.{roomId}.meta` | `room.{roomId}.meta` |
| `space.{spaceId}.joined` | `joined` |
| `space.{spaceId}.left` | `left` |
| `space.{spaceId}.member_deleted` | `member_deleted` |
| `live.instance.user.{userId}.*` | `live.user.{userId}.*` |
| `live.instance.space.{spaceId}.*` | `live.server.*` |
| `live.space.{spaceId}.*` | `live.server.*` |
| `live.space.{spaceId}.room.{roomId}.*` | `live.server.room.{roomId}.*` |

The two-tier real-time event split from [ADR-012](ADR-012-two-tier-realtime-events.md) is preserved: per-user-private under `live.user.{userId}.*`, server-wide and per-room under `live.server.*`. Wildcards remain the same shape.

### Key shapes inside `SERVER`

```
user.{id}                              # user record
auth.{id}.password                     # password hash
user.{id}.avatar                       # avatar metadata
user.{id}.verified_emails              # verified email list
user_preferences.{id}                  # per-user preferences
user_by_login.{login}                  # secondary index: login → user ID
user_by_email.{hash}                   # secondary index: email hash → user ID
member.{userId}                        # server-membership record
room.{id}                              # room record (includes type: channel|dm)
room_membership.{userId}.{roomId}      # per-room membership
config.server                          # server config (name, description, etc.)
branding.logo                          # branding metadata pointer (object lives in ASSETS)
branding.banner                        # branding metadata pointer
rbac.role.{name}                       # role definition
rbac.role_permission.{role}.{perm}     # role grant/deny
rbac.role_assignment.{role}.{userId}   # role membership
rbac.user_permission.{userId}.{perm}   # per-user grant
rbac.user_permission_denied.{userId}.{perm}  # per-user deny
rbac.first_admin_assigned              # bootstrap flag
```

### Key shapes inside `RUNTIME`

All entries use per-key TTL. The bucket is wipeable: losing it costs a forced re-login and re-fetched caches, never user data.

```
auth_token.{token}                     # bearer token
notification.{userId}.{notifId}        # notification entry
link_preview.{urlHash}                 # link preview cache
asset_cache.{attachmentId}.{paramsHash}  # transformed-image cache
room_read_status.{userId}.{roomId}     # last-read marker
room_mention_status.{userId}.{roomId}  # mention bitmap
email_verification.{token}             # email-verification token
password_reset.{token}                 # password-reset token
account_deletion.{token}               # account-deletion token
```

### Operational improvements bundled with the new layout

These are not strictly part of the schema but are landed alongside it because the new buckets are being created from scratch — there's no upgrade-in-place to worry about:

- **NATS ≥ 2.14**: pin the embedded NATS dependency to ≥ 2.14 for async stream snapshots (better R3 tail latency on backup).
- **`MaxBytes` on `PRESENCE`**: explicit memory ceiling. Memory-backed buckets without this footgun the operator if presence traffic spikes.
- **`Duplicates: 0` on `RUNTIME`**: the duplicates dedup window isn't useful for cache/token traffic and consumes RAM. Disable it.
- **Per-key TTL standardized on transactional tokens**: every entry under `RUNTIME` has an explicit TTL; nothing relies on bucket-level retention.

## Consequences

- **Backup story simplifies**. `chatto backup` enumerates a known-finite set of streams instead of discovering arbitrary `SPACE_{id}_*` bucket families. The skip list (per [`backup-restore.md`](../../.claude/rules/backup-restore.md)) shrinks: `RUNTIME`, `PRESENCE`, and `ENCRYPTION_KEYS` are skipped; everything else is backed up.
- **Restore story simplifies**. The destination buckets are a known set; restore creates exactly seven buckets, one stream, one object store. No per-space iteration during restore.
- **Stream filters require care**. The `CHAT_EVENTS` filter is explicit (`room.>`, `joined`, `left`, `member_deleted`) so that future subject namespaces (`live.>`, `audit.>`, `presence.>`) cannot accidentally land in the durable stream. A typo here is a silent disaster — covered explicitly in Phase 2 PR 6 testing.
- **Subject parsers become simpler**. Today's parsers strip a `space.{id}.` prefix; new parsers don't. The function `IsThreadSubject` still works on the `.replies.` semantic marker (per [`nats-subjects.md`](../../.claude/rules/nats-subjects.md)). `ParseRoomIDFromSubject` becomes a partial match: subjects starting with `room.` parse `parts[1]` as the room ID; other subjects (`joined`, `left`, etc.) return a sentinel.
- **DM rooms live in the same buckets as channels**. There is no `SPACE_DM_*` carve-out (see [ADR-030](ADR-030-dm-as-room.md)). DM `room.{id}` records, DM bodies, DM reactions, DM threads all live in the seven shared buckets, distinguished only by `room.type == dm`.
- **`RUNTIME` is intentionally lossy**. Wiping `RUNTIME` is a supported recovery action: users get logged out, caches refill, read markers reset to "everything unread." Operators can wipe it without losing structural data. Migrations and backup-restore explicitly do not preserve `RUNTIME` content.
- **`PRESENCE` is intentionally non-durable**. The bucket is in-memory; restart loses it. Presence reconverges from clients within seconds; this is a feature, not a bug.
- **`ENCRYPTION_KEYS` keeps its security posture**. Backed-up data without keys is unreadable; key export/import (`chatto keys export | chatto keys import`) is a separate, explicitly-prompted ceremony per [`backup-restore.md`](../../.claude/rules/backup-restore.md).
- **Cross-space iteration disappears as a concept**. Today's "for each space, for each room" enumeration becomes "for each room." Code that iterated over `ListSpaces()` is rewritten or deleted. Cross-space queries that were "expensive" per [ADR-013](ADR-013-per-space-stream-sharding.md) are now just "across the server" and not particularly expensive.
- **Lazy initialization disappears**. There is no per-space resource that needs to be ensured on first use. The `ensureSpaceStream` `sync.Map` cache and the per-space `lazycache.Cache[jetstream.KeyValue]` go away. Buckets and streams are created once at server startup. This trades a one-time small startup cost for the cognitive benefit of "everything is always there."
- **Migration is a once-off, not an ongoing concern**. The destination shape is the only shape the running server knows; there are no compatibility shims for the old layout. Phase 4's `chatto migrate` is the only place old → new translation happens, and it runs once per deployment.

## References

- Closes [#291](https://github.com/chattocorp/chatto/issues/291).
- Supersedes the per-space sharding aspect of [ADR-013](ADR-013-per-space-stream-sharding.md). The KV-as-source-of-truth/streams-as-audit-log split from [ADR-006](ADR-006-kv-source-of-truth-streams-audit-log.md) carries over unchanged.
- Builds on [ADR-007](ADR-007-per-user-encryption-with-crypto-shredding.md) (compound BODIES key for crypto-shredding), [ADR-011](ADR-011-message-body-event-split.md) (body / event split), [ADR-012](ADR-012-two-tier-realtime-events.md) (live subject taxonomy), [ADR-016](ADR-016-occ-for-message-publishing.md) (THREADS optimistic locking).
- Companion to [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md), [ADR-027](ADR-027-user-identity-and-oidc-binding.md), [ADR-028](ADR-028-permission-model-post-merge.md), [ADR-030](ADR-030-dm-as-room.md).
