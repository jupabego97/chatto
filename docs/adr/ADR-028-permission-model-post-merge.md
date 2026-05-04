# ADR-028: Permission Model After the Server Consolidation

**Date:** 2026-05-04

## Context

Today's Chatto runs two parallel RBAC systems: an *instance* layer (`instance-owner`, `instance-admin`, `instance-moderator`, `everyone`) and a *space* layer (`owner`, `admin`, `moderator`, `everyone`). Permission strings are similarly bifurcated across three scopes (`instance`, `space`, `room`). Both layers use the same `rbac.Engine` machinery and the same hierarchy-wins resolution from [ADR-005](ADR-005-hierarchy-wins-rbac.md), but they're configured separately, surfaced through separate admin UIs, and checked through parallel `Can*` helpers in the GraphQL boundary.

[ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) collapses the Instance + Space layers into one Server. With one membership boundary instead of two, there is exactly one role list to maintain and exactly two scopes worth distinguishing: server-wide (applies everywhere) and room-level (overrides for a specific room).

This ADR locks the post-merge permission model: which roles exist, which scopes exist, and which permissions are kept, renamed, or dropped.

## Decision

### One role namespace

Four built-in system roles, in rank order (lowest position = highest rank):

1. `owner`
2. `admin`
3. `moderator`
4. `everyone` (virtual; every user is implicitly a member)

The `instance-` prefix is gone. The two parallel role lists become this single list. Operators can create additional custom roles, ranked anywhere in the hierarchy, exactly as today.

### Two scopes

Permissions live in one of two scopes:

- **`server`**: applies across the entire server (e.g. `server.manage`, `member.invite`, `dm.view`).
- **`room`**: applies to a specific room and may override the server-level permission for that role in that room (e.g. denying `message.post` on `everyone` in an announcements room while leaving the server-level `message.post` grant intact for higher-rank roles).

The previous three-way split (`instance` / `space` / `room`) collapses to this two-way split. Internally the resolver still walks "server scope first, then room-level overrides" — that's the same hierarchy as today, minus one tier.

### Hierarchy-wins resolution carries over

Resolution behavior is exactly as described in [ADR-005](ADR-005-hierarchy-wins-rbac.md): for a given user and permission, walk the user's roles by rank (highest first); the first role with an explicit grant or deny wins; default-deny if no role has an opinion. Per-user overrides (direct grants and direct denies) take precedence over role-derived decisions. Room-level overrides take precedence over server-level decisions for the same role.

### Permission rename / drop / keep table

The full mapping from today's permissions to the post-merge set:

| Today | Action | New name | Reason |
|---|---|---|---|
| `space.list` | DROP | — | No in-app server discovery; one server per process. |
| `space.create` | DROP | — | No nested servers. |
| `space.join` | DROP | — | Joining the server is registering an account; not a separate gated action. |
| `space.leave` | RENAME | `server.leave` | Same semantic, server-scoped. |
| `space.manage` | RENAME | `server.manage` | Server settings (name, description, branding) edit access. |
| `space.delete` | DROP | — | Server lifecycle is operator-controlled (process management), not an in-app permission. |
| `admin.view-spaces` | DROP | — | No spaces admin page. |
| `admin.access`, `admin.users-view`, `admin.users-manage`, `admin.roles-view`, `admin.roles-manage`, `admin.system-view`, `admin.audit-view` | KEEP | unchanged | Audit individually after rename; no semantic change in this ADR. |
| `dm.view`, `dm.write` | KEEP | unchanged | Global server-scoped toggles for whether DMs are usable at all. See [ADR-030](ADR-030-dm-as-room.md). |
| `room.list`, `room.create`, `room.join`, `room.leave`, `room.manage` | KEEP | unchanged | Already correctly room/server-scoped. |
| `message.post`, `message.post-in-thread`, `message.reply`, `message.reply-in-thread`, `message.edit-own`, `message.edit-any`, `message.delete-own`, `message.delete-any`, `message.react`, `message.echo` | KEEP | unchanged | Rescope from `space` → `server` for the default scope; per-room overrides still work. |
| `member.invite`, `member.remove` | KEEP | unchanged | Server-scoped. |
| `role.manage`, `role.assign` | KEEP | unchanged | Server-scoped. |
| `user.delete`, `user.delete-self` | KEEP | unchanged | Server-scoped. |

The `instance-*` permissions in the old model that mirror the kept entries above (e.g. an instance-level `member.invite`) are removed by virtue of the rename — there is now one `member.invite` permission, not two. Where instance-level and space-level permissions today carry the same name in their respective scopes, they collapse to a single permission in the unified scope.

The full list of post-merge permissions is the source of truth for the implementation in `cli/internal/core/permissions.go` after the Phase 2 refactor lands; this table is the design lock, not a duplicate manifest.

### `owners.emails` short-circuit unchanged

The `owners.emails` config block continues to designate server owners by verified-email match (per [ADR-027](ADR-027-user-identity-and-oidc-binding.md)). The matching short-circuits the role hierarchy: a user matching `owners.emails` is treated as `owner` with all permissions granted, bypassing role lookup entirely. The mechanics are identical to today's instance-owner short-circuit; only the framing ("server" instead of "instance") changes.

### First-admin bootstrap unchanged

The first user to sign in to a freshly-installed server still becomes `owner` automatically, exactly as today's first-instance-admin bootstrap. The `rbac.first_admin_assigned` flag continues to mark the bootstrap as completed.

## Consequences

- **Single admin UI for roles and permissions**. The two parallel admin pages (`/admin/roles` for instance, `/[spaceId]/admin/roles` for space) collapse to one. The permission inspector simplifies — one scope hierarchy instead of two.
- **Roughly half the RBAC code disappears**. `instance_can.go` + `space_can.go` merge into `can.go`. `instance_rbac.go` + `space_rbac.go` collapse. The two parallel sets of GraphQL `viewerCan*` resolvers become one.
- **Scope reduction is observable to advanced operators**. Anyone who's manually inspected RBAC KV keys today sees three scopes; after the merge they see two. The KV key format from the rbac package (`{role}:{allow|deny}:{verb}.{objectType}:{objectId}`) is unchanged in shape; only the populated `objectType` set changes.
- **No customization of DM permissions beyond the global toggles**. `dm.view` and `dm.write` are global server-scoped permissions; per-room DM-permission overrides aren't a thing. This matches the post-consolidation reality that DMs are regular rooms (see [ADR-030](ADR-030-dm-as-room.md)) but with their permission story expressed at the server scope rather than the room scope.
- **Drops are real drops, not deprecations**. `space.list`, `space.create`, `space.join`, `space.delete`, `admin.view-spaces` are removed entirely; permission-grants referring to them in existing data are dropped during the migration ([Phase 4 migration ADR — to be written](https://github.com/chattocorp/chatto/issues/284)). There is no compatibility shim.
- **Custom roles defined by operators carry over by name**. A custom role named `gardener` keeps its name and position. Its permission grants are inspected against the new permission set during migration; grants referring to dropped permissions are silently removed (logged in the migration report).
- **Tests need updating, not redesigning**. The hierarchy-wins property tests from [ADR-005](ADR-005-hierarchy-wins-rbac.md) carry over verbatim against the unified role list. The DM-permission carve-out (`isDMPermissionAllowed` hardcoded allow-list) is removed; DM permissions now resolve through the unified engine plus the `dm.view`/`dm.write` toggles.

## References

- Closes [#292](https://github.com/chattocorp/chatto/issues/292).
- Builds on [ADR-005](ADR-005-hierarchy-wins-rbac.md) (Hierarchy-wins RBAC).
- Companion to [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) (Server consolidation), [ADR-027](ADR-027-user-identity-and-oidc-binding.md) (identity), [ADR-029](ADR-029-server-destination-schema.md) (destination schema), [ADR-030](ADR-030-dm-as-room.md) (DM-as-room).
