# Permissions

Chatto uses a unified RBAC model: every permission is defined once and can be configured at three tiers (instance / space / room). The same resolution algorithm runs at every tier, so the rules an operator learns in one place hold everywhere.

## The rule

> Walk the user's applicable roles in hierarchy order (lower position = higher rank). For each role, look up its decision at the current tier; if the role has no entry there, walk **upward** along that role's tier chain. The first role to produce **any** explicit decision wins. If no role decides anywhere, the answer is **deny**.

Two practical consequences:

- **Higher-rank roles always override lower-rank roles**, regardless of which tier the decision lives in.
- **A grant cascades down by default.** Granting `message.post` to `instance-everyone` at instance scope makes every authenticated user able to post in every room they're a member of, without per-space configuration.

The implementation lives in `cli/internal/core/permission_resolver.go` (`PermissionResolver.walk`).

## Role hierarchy

| Position | Instance role        | Space role  | Purpose |
|----------|----------------------|-------------|---------|
| 0        | `instance-owner`     | `owner`     | Top of the hierarchy. Carries every permission. |
| 1        | `instance-suspended` | `suspended` | Carries explicit denies on user-behavior permissions. Outranks admin so the denies stick — assign this role and the user can't post, react, join rooms, etc. |
| 2        | `instance-admin`     | `admin`     | Full admin power at its tier. At instance scope this means every permission everywhere. |
| 3        | —                    | `moderation`| Heavy moderation powers (`message.delete-any`, `member.remove`). Designed to be granted on demand rather than baked into the moderator badge. |
| 4        | `instance-moderator` | `moderator` | Light moderation badge. By default: `room.manage` at space scope, read-only admin at instance scope. |
| 5+       | custom roles                       || Always rank below moderator. |
| MaxInt32 | `everyone`           | `everyone`  | Implicit floor role for every authenticated user / space member. |

System roles are recreated and migrated to their canonical positions on every boot (`initInstanceRBAC` and `migrateSpaceSystemRoles`). Existing per-space role configurations are preserved — only the role definitions themselves are normalized.

## Worked examples

### Suspending a user

A staff member assigns `instance-suspended` to a problem user.

- The role sits at position 1, above `instance-admin` (2) — but no admin is involved here, the user just has `instance-suspended` + `instance-everyone`.
- `instance-suspended` carries explicit denies for `space.create`, `space.join`, `dm.view`, `dm.write`, `admin.access`, `user.delete-self` at instance scope, and `message.{post,react,...}` + `room.{create,join}` at space scope (via the per-space `suspended` role).
- When the user tries to post a message, the resolver walks `instance-suspended` first (highest rank). At room scope it has no entry, walks up to space, walks up to instance — at instance there's a deny on `message.post`'s siblings but not on `message.post` itself; however the per-space `suspended` role *does* have a deny at space scope. Either way, the deny wins before any lower-rank role gets to grant.
- The user can still leave the space (`space.leave` isn't denied), so they're not stuck.

### Moderation on demand

A space admin wants a user to be able to delete messages and remove members for a specific incident, but doesn't want them to keep those powers permanently.

- The space ships with a `moderation` role at position 3 with `message.delete-any` + `member.remove` granted. By default it has zero members.
- The admin assigns `moderation` to the user.
- When the user clicks "delete message" in another user's post, the resolver walks `moderation` first. It finds an allow at space scope and stops.
- After the incident, the admin revokes the `moderation` role. The user goes back to whatever they had before — moderator badge, custom role, or just `everyone`.

### Announcement-only rooms

A space wants `#announcements` where only owner/admin/moderator can post root messages, but everyone can read and reply in threads.

- At room scope on `#announcements`, deny `message.post` to `everyone` and explicitly grant it to `owner` / `admin` / `moderator` (handled by `SetupAnnouncementsRoomPermissions` in core).
- When `everyone`-only user tries to post a root message, the resolver walks `everyone` (their only applicable role for this perm). At room scope it finds the deny. Done.
- When an admin posts, the resolver walks `instance-admin` first. At room scope: no entry. At space scope: no entry. At instance scope: allowed by `DefaultInstanceFullAllows`. Allowed.
- Everyone can still post in threads because `message.post-in-thread` is on `instance-everyone` and not denied at the room.

## Where to look in the code

- **Permission catalog**: `cli/internal/core/permission.go`. Edit `allPermissions` and the corresponding `PermXxx` constant when adding a new permission.
- **Default seeds**: same file, the `Default*` functions, plus the `InitInstanceDefaults` / `InitSpaceDefaults` wiring in `permission_ops.go`.
- **Resolution**: `cli/internal/core/permission_resolver.go`. The `walk` function is the only place permission decisions are made.
- **Inspector**: `cli/internal/graph/permission_inspector*.go` exposes the trace as the `permissionExplanation` GraphQL query. The frontend `PermissionInspectorPanel` renders it.
- **Matrix UI**: `frontend/src/lib/components/rbac/PermissionMatrix.svelte` is the per-tier editor.
