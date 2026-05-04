# ADR-030: Direct Messages as Regular Rooms

**Date:** 2026-05-04

## Context

[ADR-015](ADR-015-dms-as-hidden-space.md) modeled direct messages as rooms inside a hardcoded synthetic space (`spaceID = "DM"`). That decision bought meaningful infrastructure reuse — DMs went through the same JetStream stream, the same body storage, the same subscription fan-in — but it also created a long tail of carve-outs: `IsDMSpace` checks scattered through ~15 call sites, a bypass of the RBAC engine for DM permissions (`isDMPermissionAllowed`), filtering at the GraphQL boundary so the synthetic space wasn't visible in `Query.spaces`, and a special `requireSpaceMember` path for DM rooms.

[ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) explicitly identifies this DM workaround as one of the most visible places the two-layer model leaks, and it makes folding DMs into the regular room model one of the three coupled parts of the consolidation. With per-space buckets gone (per [ADR-029](ADR-029-server-destination-schema.md)) and the DM space's bucket family disappearing along with them, the synthetic-space-as-storage-host idea is no longer tenable. The natural shape is "DMs are rooms with a `type` field."

This ADR locks the DM-as-room model: how DM rooms are distinguished, what stays the same as today, what's deferred, and how the existing find-or-create semantics survive the change.

## Decision

### Two room types

The `Room` proto gets a `type` enum field with two values:

| Type | Meaning |
|---|---|
| `channel` | Named, listed in the room directory, joinable per `room.join` permission. |
| `dm` | Hidden from the room directory, listed via a dedicated DM-list query. ID is deterministic from participants. 2 to 10 participants. |

The `type` field defaults to `channel` for backward compatibility on records that pre-date the field (relevant only briefly during Phase 2; the migration in Phase 4 sets `type` explicitly on every record).

There is **no `visibility` field** on rooms in this consolidation. Private channels (channels hidden from non-members) are tracked separately and explicitly deferred. DM hiding is achieved by filtering on `type`, not by a separate visibility flag.

### DM rooms live in the same buckets as channel rooms

Per [ADR-029](ADR-029-server-destination-schema.md):

- DM `room.{id}` records live in the `SERVER` bucket alongside channel rooms.
- DM messages flow through the single `CHAT_EVENTS` stream with subject `room.{roomId}.msg.{eventId}`, identical to channel messages.
- DM bodies live in `BODIES` keyed `{userId}.{eventId}`. DM reactions live in `REACTIONS` keyed `{roomId}.{eventId}.{emoji}`. DM thread metadata lives in `THREADS` keyed `{roomId}.{rootEventId}`.
- DM room memberships live in `SERVER` keyed `room_membership.{userId}.{roomId}`, identical to channel memberships.

There is no `SERVER_DM_*` carve-out, no synthetic DM-only stream, no DM-only bucket. The previous `SPACE_DM_*` family disappears entirely.

### DM room IDs remain deterministic

DM room IDs continue to be computed as the first 14 hex characters of `SHA-256(sorted participant user IDs joined by ".")`, exactly as in [ADR-015](ADR-015-dms-as-hidden-space.md). This preserves the find-or-create coordination-free property: two participants concurrently starting a DM compute the same room ID and converge on the same room without locking.

DM room IDs are unprefixed (no `R` prefix per [ADR-022](ADR-022-nanoid-with-entity-prefixes.md)) because they are computed, not random. The 14-character length matches the `R{14-char}` length of channel rooms, so all room IDs have the same visual footprint.

### Permissions: two global toggles, no per-room overrides

Per [ADR-028](ADR-028-permission-model-post-merge.md), DM permissions are two global server-scoped permissions:

| Permission | Meaning |
|---|---|
| `dm.view` | Whether the user can see DM rooms at all. |
| `dm.write` | Whether the user can post into DM rooms. |

These are **not** per-room permissions. There is no support for "this user can DM Alice but not Bob," or "this DM room has different posting rules." The toggles are server-wide:

- A user with `dm.view` denied sees no DM list and cannot open DM rooms.
- A user with `dm.write` denied can read DM rooms they're a member of (history) but cannot post.
- An admin removing `dm.write` from `everyone` is the supported "disable DMs for regular users" pattern.

For everything else (membership, message visibility, edit/delete-own, reactions), DM rooms resolve through the normal RBAC engine on the unified server scope. The hardcoded `isDMPermissionAllowed` allow-list from the old model is removed.

### Discovery semantics

| Surface | Filter |
|---|---|
| `Query.rooms` (room directory / browse) | `type == channel` only. |
| `Query.dmConversations` (DM list) | `type == dm` only. Returns DM rooms the caller is a member of. |
| Room URL deep-link | Either type, but the user must be a member to see content. |

The dedicated DM-list query is necessary because DM membership is per-user (you only see DMs you're in), whereas channel discovery is server-wide (you see channels regardless of whether you've joined). The two queries take different paths through the resolver.

### Out of scope — explicitly deferred

The following are **not** part of this ADR. Each can be added as an additive change on top of the DM-as-room model without breaking the design:

- **Named group DMs.** A group DM could carry an optional human-set name. The `type` enum would either gain a `group` value or the `dm` type would gain an optional `name` field. Tracked as part of "DMs 2.0" in a future ADR.
- **Mutable group DM membership.** Today's group DM membership is fixed at creation (the deterministic ID encodes the participant set). A "DMs 2.0" feature could allow add/remove of participants, at which point the room ID stops being a hash of participants and becomes a regular `R…` NanoID.
- **Channel ↔ group DM promotion.** Converting a small private channel to a group DM (or vice versa) is conceivable; not in scope.
- **Distinguishing 1:1 from multi-person DMs as separate types.** They share the `dm` type today; a future split into `dm` (1:1) and `group` (3+) is possible without breaking the model.
- **Private channels.** A `visibility: private` channel hidden from non-members is a separate feature. Tracked elsewhere.
- **Per-DM-room permission overrides.** Not supported. If a future feature requires it (e.g. read-only DMs for verified safety reasons), a separate ADR will be needed.

## Consequences

- **Symmetric codebase**. `IsDMSpace` and the `"DM"` constant disappear. Every call site that today branches on "is this a DM?" either drops the branch (the unified RBAC and storage paths handle both) or substitutes `room.Type == RoomTypeDM` for the rare cases where the rendering or filtering genuinely differs.
- **Unified RBAC**. DM permission resolution goes through the same engine as channel permission resolution, modulated by the global `dm.view`/`dm.write` toggles. The `isDMPermissionAllowed` carve-out is removed.
- **Find-or-create unchanged**. Two participants concurrently invoking `startDM` still converge on the same room without coordination, because the deterministic SHA-256 hash of sorted participant IDs hasn't changed.
- **Group DMs continue to work**. The deterministic ID is built from a sorted *set* of participant IDs, so 3-to-10-participant DMs work with the same mechanism as 1:1 DMs. The 10-participant cap from today's model is preserved.
- **Operators can disable DMs server-wide**. Removing `dm.view` and `dm.write` from `everyone` (or from any role) is now an in-scope, supported configuration. The old "DM permissions are hardcoded, no customization" limitation from [ADR-015](ADR-015-dms-as-hidden-space.md) is lifted at the on/off level — though per-DM-room customization remains out of scope.
- **DM history is reset at the cutover migration**. Per [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md), the Chatto Community migration drops DM data and only carries forward channel data. New DM rooms get created in the new model on first-DM after cutover; users start DMs fresh.
- **GraphQL surface**. `Mutation.startDM` continues to exist (creating a `type: dm` room). `Query.dmConversations` is the dedicated listing endpoint. The DM list at `/chat/dm/...` on the frontend continues to live under that route — the URL doesn't change just because the storage model did, since DMs remain a meaningfully distinct user-facing surface.
- **Future DM features are additive**. Naming groups, mutable membership, and split 1:1-vs-group types can land later without touching the storage layout, the stream, or the RBAC engine. The `type` field is the extension point.

## References

- Closes [#293](https://github.com/chattocorp/chatto/issues/293).
- Supersedes [ADR-015](ADR-015-dms-as-hidden-space.md) (DMs as a hidden space), which is also superseded by the parent [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md).
- Builds on [ADR-022](ADR-022-nanoid-with-entity-prefixes.md) (DM room ID as deterministic SHA-256).
- Companion to [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md), [ADR-027](ADR-027-user-identity-and-oidc-binding.md), [ADR-028](ADR-028-permission-model-post-merge.md), [ADR-029](ADR-029-server-destination-schema.md).
