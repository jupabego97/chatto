# Authorization Model

This document describes the authorization requirements and policies for Chatto's GraphQL API.

## Core Principles

1. **Users are bound to an instance** - All users exist within a single Chatto instance
2. **Spaces are discoverable** - Users can browse all spaces for discovery purposes. **Anonymous callers always see all spaces** regardless of any `space.list` permission configuration on the `everyone` role; revoking `space.list` only affects authenticated non-members. To make an instance fully private, place it behind a reverse-proxy auth layer or disable the `Query.spaces` / `Query.space(id)` resolvers via a future configuration flag — the GraphQL gateway alone cannot enforce private-instance discovery.
3. **Room access requires space membership** - Users must join a space before accessing its rooms
4. **Message access requires room membership** - Users can only read/write messages in rooms they've joined
5. **User profiles are public** - Basic user info (id, login, displayName, avatar) is visible to all authenticated users
6. **Membership info is private** - Users can only see their own space/room memberships

## Authorization Architecture

Authorization is enforced at the **API boundary**, not in core:

| Layer | Responsibility |
|-------|----------------|
| **GraphQL** | User-facing API. Checks authorization via `Can*` functions before calling core. |
| **Core** | Pure business logic. Assumes caller is authorized. Documents requirements in comments. |
| **NATS** | Extension/internal API. Trusted context, calls core directly. |

**Why this design:**
- Core functions are reusable from trusted contexts (NATS handlers, background jobs)
- No redundant permission checks when core calls other core functions
- Clear separation: GraphQL handles user authorization, core handles business logic
- Audit logging can be added orthogonally without coupling to authorization

## Permission System

Permissions are granted through roles assigned to instance users and space members. Use the `Can*` functions in `core/instance_can.go` and `core/space_can.go`, or check directly via `core.HasInstancePermission` / `core.HasSpacePermission` / `core.HasRoomPermission`.

### Harmonized Resolution Model

Resolution is the **same algorithm at every tier** (`PermissionResolver.walk` in `permission_resolver.go`):

1. Walk the user's applicable roles in hierarchy order (lower position = higher rank).
2. For each role, look up its decision at the requested tier; if no entry exists, walk **upward** (room → space → instance) along that role's chain.
3. The first role to produce **any** explicit decision (allow or deny) wins.
4. If no role decides anywhere, the answer is deny.

Two consequences worth internalizing:

- **Higher-rank roles always override lower-rank roles**, regardless of which tier the decision lives in. An `instance-admin` allow at instance scope beats an `everyone` deny at space scope, because admin is checked first.
- **A role's grants cascade down by default.** Granting `message.post` to `instance-everyone` at instance scope is visible inside every space the user joins until something at a lower tier (or higher-rank role at any tier) overrides it.

### Role Hierarchy

| Position | Instance role        | Space role  | Notes |
|----------|----------------------|-------------|-------|
| 0        | `instance-owner`     | `owner`     | All permissions; cannot be denied |
| 1        | `instance-suspended` | `suspended` | Carries explicit denies on user-behavior perms; outranks admin so denies stick |
| 2        | `instance-admin`     | `admin`     | All permissions at instance / space-management at space |
| 3        | —                    | `moderation`| Heavy moderation grants (delete-any, member.remove); space scope only |
| 4        | `instance-moderator` | `moderator` | Read-only admin / room.manage |
| 5+       | custom roles                       || Custom roles always rank below moderator |
| MaxInt32 | `everyone`           | `everyone`  | Universal floor role |

System roles get their canonical positions from constants in `cli/internal/core/rbac/engine.go` and are migrated automatically on every boot (`initInstanceRBAC` and `migrateSpaceSystemRoles`).

**Testing implication:** Denying a permission on a low-rank role (like `everyone`) does **not** block users with a higher-rank role unless that role doesn't have an explicit allow. To reliably block a permission, deny on a role that ranks at or above any role that grants it (e.g. `instance-suspended`).

### Permission Constant Naming

Permission constants follow the pattern `InstPerm{Category}{Action}` (singular nouns):

| Pattern | Example | Notes |
|---------|---------|-------|
| `InstPerm{Category}{Action}` | `InstPermSpaceCreate` | Singular category |
| `InstPermAdmin{Area}{Action}` | `InstPermAdminUsersView` | Admin permissions |
| `InstPermDM{Action}` | `InstPermDMWrite` | DM permissions |

**Common mistakes** (avoid these):
- `InstPermSpacesCreate` → Use `InstPermSpaceCreate` (singular)
- `InstPermDMsWrite` → Use `InstPermDMWrite` (no plural 's')
- `InstPermAdminAccessUsersView` → Use `InstPermAdminUsersView`

The Go constants in `cli/internal/core/permissions.go` are the source of truth. Frontend TypeScript types are generated via `mise codegen-types`.

### Permission String Naming

Permission strings use **hyphens** as word separators (e.g., `message.post-in-thread`, `message.edit-own`, `message.reply-in-thread`). Never use underscores in permission strings.

### Built-in Permissions

The full list lives in `cli/internal/core/permission.go` (`allPermissions`). Highlights:

| Permission | Description |
|------------|-------------|
| `space.manage` | Update space settings (name, description) |
| `space.delete` | Delete the space |
| `role.manage` | Create/edit/delete roles |
| `role.assign` | Assign roles to users |
| `member.invite` | Invite new members |
| `member.remove` | Remove members from space (lives on the `moderation` role by default) |
| `room.list` | View the list of rooms |
| `room.create` | Create new rooms |
| `room.join` | Join existing rooms |
| `room.manage` | Update/delete any room |
| `message.delete-any` | Delete any user's messages (lives on the `moderation` role by default) |

**Every permission can be configured at instance scope.** A grant or deny at instance scope propagates down through the resolver's parent-tier fallback to every space and room (subject to membership). Space- and room-scoped configuration overrides the inherited baseline for that specific scope.

## GraphQL Authorization Reference

### Queries

| Query | Auth Required | Additional Check |
|-------|---------------|------------------|
| `me` | No | Returns null if unauthenticated |
| `user(id)` | No | Public user profiles |
| `users` | Yes | Instance admin only |
| `spaces` | No | Discovery - lists all spaces |
| `space(id)` | No | Discovery - view any space |
| `room(spaceId, roomId)` | Yes | Room membership required |
| `roomEvents(...)` | Yes | Room membership required |
| `roomEvent(...)` | Yes | Room membership required |
| `admin` | Yes | Instance admin only |

### Mutations

| Mutation | Auth Required | Additional Check |
|----------|---------------|------------------|
| `createUser` | No | Self-registration |
| `createSpace` | Yes | None (anyone can create) |
| `updateSpace` | Yes | `space.manage` |
| `joinSpace` | Yes | None |
| `leaveSpace` | Yes | None |
| `createRoom` | Yes | `rooms.create` |
| `joinRoom` | Yes | Space membership + `rooms.join` |
| `leaveRoom` | Yes | None |
| `postMessage` | Yes | Room membership + `message.post` (root) or `message.post-in-thread` (thread reply), + `message.reply` (if `inReplyTo` in room) or `message.reply-in-thread` (if `inReplyTo` in thread), + `message.echo` (if `alsoSendToChannel`) |
| `markRoomAsRead` | Yes | Room membership |
| `addReaction` | Yes | Room membership |
| `removeReaction` | Yes | Room membership |
| `deleteMessage` | Yes | Room membership + message ownership |
| `updateMyPresence` | Yes | None (sets caller's own presence) |

### Subscriptions

| Subscription | Auth Required | Additional Check |
|--------------|---------------|------------------|
| `mySpaceEvents(spaceId)` | Yes | Space membership |
| `mySpaceLiveEvents(spaceId)` | Yes | Space membership |
| `myInstanceEvents` | Yes | None (user's own events) |
| `presenceUpdates(spaceId)` | Yes | Space membership |

### Field Resolvers

| Field | Auth Required | Additional Check |
|-------|---------------|------------------|
| `Space.rooms` | Yes | Space membership + `rooms.browse` |
| `Space.memberCount` | No | Public count |
| `Space.roomCount` | No | Public count |
| `Space.assetCount` | No | Public count |
| `Room.members` | Yes | Room membership |
| `Room.hasUnread` | No | Returns false if unauthenticated |
| `User.spaces` | Yes | Self only (`caller.Id == obj.Id`) |
| `User.rooms` | Yes | Self only (`caller.Id == obj.Id`) |
| `User.avatarURL` | No | Public |
| `User.presenceStatus` | No | Public |

## Implementation Patterns

### GraphQL Resolver with Permission Check
```go
func (r *mutationResolver) CreateRoom(ctx context.Context, input model.CreateRoomInput) (*Room, error) {
    user, err := requireAuth(ctx)
    if err != nil {
        return nil, err
    }

    // Check permission at GraphQL layer
    can, err := r.core.CanCreateRoom(ctx, user.Id, input.SpaceID)
    if err != nil {
        return nil, err
    }
    if !can {
        return nil, core.ErrPermissionDenied
    }

    // Core function assumes caller is authorized
    return r.core.CreateRoom(ctx, user.Id, input.SpaceID, input.Name, input.Desc)
}
```

### Core Function (no authorization check)
```go
// CreateRoom creates a new room in a space.
// Authorization: Caller must verify CanCreateRoom before calling.
func (c *ChattoCore) CreateRoom(ctx context.Context, actorID, spaceID, name, desc string) (*Room, error) {
    // Business logic only - no permission check here
}
```

### Authentication Helpers (in graph/authz.go)
```go
user, err := requireAuth(ctx)           // Returns authenticated user or error
user, err := requireSpaceMember(ctx, r.core, spaceID)  // + space membership
user, err := requireRoomMember(ctx, r.core, spaceID, roomID)  // + room membership
```

### Self-Only Access Check
```go
if caller.Id != obj.Id {
    return nil, fmt.Errorf("access denied: cannot view other users' data")
}
```

## Customizable Permissions

Default permissions are seeded by `InitInstanceDefaults` and `InitSpaceDefaults` (in `permission_ops.go`) — including the user-behavior floor on `instance-everyone`, the moderation grants on `space.moderation`, and the suspended denies on `*-suspended`. All of them can be granted, denied, or cleared at any tier through the matrix UI. Two rules to keep in mind when adding permission checks:

1. **Always go through `PermissionResolver`.** Never hardcode "this role implicitly has X" — admins are not exempt from `everyone`-role denies, only out-ranked by them.
2. **Test both grant and revoke at the actually-effective tier.** A test that denies `message.post` on `everyone` will not block an `admin` user; deny on the user's highest-rank role (or use the `suspended` role) instead.

## Instance Admin

Instance admins are configured via `admin.emails` in `chatto.toml`. They have access to:

- `/admin` routes in the frontend
- `Query.admin` and `Query.users` in GraphQL
- System monitoring data (NATS stats, streams, KV buckets)

Instance admin is separate from space admin roles - see `admin.md` for details.
