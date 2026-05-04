package core

import "context"

// This file provides semantic helper functions wrapping the low-level
// HasInstancePermission / hasSpacePermission / hasRoomPermission calls with
// business-meaningful names. Each function returns (bool, error) where
// `error` is non-nil only on a system error.
//
// Per ADR-028 the post-merge permission model has just two scopes (`server`
// and `room`); the dual-RBAC engines collapse in PR 4 of the Phase 2 refactor.
// Until then this file is the unified surface.
//
// Note: these helpers check RBAC permissions only. Config-based admin checks
// (owners.emails) are layered on by the caller, typically at the GraphQL
// resolver boundary.

// ============================================================================
// Server-level permissions (formerly instance-level)
// ============================================================================

// CanAdminAccess checks if a user can access the admin panel.
func (c *ChattoCore) CanAdminAccess(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermAdminAccess)
}

// CanAdminUsersView checks if a user can view the users page in admin.
func (c *ChattoCore) CanAdminUsersView(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermAdminUsersView)
}

// CanAdminUsersManage checks if a user can edit user role assignments.
func (c *ChattoCore) CanAdminUsersManage(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermAdminUsersManage)
}

// CanAdminRolesView checks if a user can view the roles page in admin.
func (c *ChattoCore) CanAdminRolesView(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermAdminRolesView)
}

// CanAdminRolesManage checks if a user can create/edit instance roles and their permissions.
func (c *ChattoCore) CanAdminRolesManage(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermAdminRolesManage)
}

// CanAdminSystemView checks if a user can view the system and data pages in admin.
func (c *ChattoCore) CanAdminSystemView(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermAdminSystemView)
}

// CanDMView checks if a user can access DM conversations and read direct messages.
func (c *ChattoCore) CanDMView(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermDMView)
}

// CanDMWrite checks if a user can start DM conversations and send messages.
func (c *ChattoCore) CanDMWrite(ctx context.Context, userID string) (bool, error) {
	return c.HasInstancePermission(ctx, userID, PermDMWrite)
}

// CanAdminManageUser checks if an actor can perform admin user-management
// actions (e.g. editing identity, clearing cooldowns) on a target user based
// on instance role hierarchy. Self-management is always allowed; otherwise
// the actor's highest role must outrank the target's highest role.
//
// Note: this checks RBAC hierarchy only. Config-based admins (owners.emails)
// are not visible to the RBAC engine and should bypass this check at the
// resolver layer (they outrank everyone).
func (c *ChattoCore) CanAdminManageUser(ctx context.Context, actorID, targetID string) (bool, error) {
	if actorID == targetID {
		return true, nil
	}
	return c.instanceRBACEngine.CanUserManageUser(ctx, actorID, targetID)
}

// CanDeleteUser checks if an actor can delete a specific user account.
// Returns true if:
//   - The actor is deleting their own account and has user.delete-self, OR
//   - The actor has the user.delete permission (admin capability)
func (c *ChattoCore) CanDeleteUser(ctx context.Context, actorID, targetUserID string) (bool, error) {
	if actorID == targetUserID {
		return c.HasInstancePermission(ctx, actorID, PermUserDeleteSelf)
	}
	return c.HasInstancePermission(ctx, actorID, PermUserDelete)
}

// ============================================================================
// Space-scoped permissions (server admin actions inside a particular space)
//
// These keep their existing names through the dual-RBAC transition. The
// CanAdminSpace* / CanSpace* / CanXxxRoom split mirrors the GraphQL surface
// and gets renamed alongside that surface in PR 10.
// ============================================================================

// spaceAdminPermissions is the set of admin-level space permissions.
// Used by HasAnyAdminPermission to check for space admin access.
//
// Per ADR-028, space.delete is dropped (server lifecycle is operator-controlled,
// not an in-app permission) and space.manage is renamed to server.manage.
var spaceAdminPermissions = []Permission{
	PermServerManage,
	PermRoleManage,
	PermRoleAssign,
	PermMemberInvite,
	PermMemberRemove,
	PermRoomManage,
}

// HasAnyAdminPermission checks if a user has any admin.* permission in a space.
// This is used to determine if the user should see the Space Admin link.
func (c *ChattoCore) HasAnyAdminPermission(ctx context.Context, userID, spaceID string) (bool, error) {
	for _, perm := range spaceAdminPermissions {
		has, err := c.hasSpacePermission(ctx, spaceID, userID, perm)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

// CanAdminSpaceManage checks if a user can update space settings (name, description, branding).
// Backed by the renamed PermServerManage permission per ADR-028.
func (c *ChattoCore) CanAdminSpaceManage(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermServerManage)
}

// CanSpaceRolesManage checks if a user can create, update, delete roles and grant/revoke permissions in a space.
func (c *ChattoCore) CanSpaceRolesManage(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermRoleManage)
}

// CanSpaceRolesAssign checks if a user can assign or revoke roles to/from other users in a space.
func (c *ChattoCore) CanSpaceRolesAssign(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermRoleAssign)
}

// CanAdminMembersInvite checks if a user can invite new members to the space.
func (c *ChattoCore) CanAdminMembersInvite(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermMemberInvite)
}

// CanAdminMembersRemove checks if a user can remove other members from the space.
func (c *ChattoCore) CanAdminMembersRemove(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermMemberRemove)
}

// CanAdminRoomsManage checks if a user can update or delete any room in the space.
func (c *ChattoCore) CanAdminRoomsManage(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermRoomManage)
}

// CanBrowseRooms checks if a user can view the list of rooms in the space.
func (c *ChattoCore) CanBrowseRooms(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermRoomList)
}

// CanCreateRoom checks if a user can create new rooms in the space.
func (c *ChattoCore) CanCreateRoom(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermRoomCreate)
}

// CanJoinRoom checks if a user can join existing rooms in the space.
func (c *ChattoCore) CanJoinRoom(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermRoomJoin)
}

// CanLeaveSpace checks if a user can leave a specific space.
// Backed by the renamed PermServerLeave permission per ADR-028.
func (c *ChattoCore) CanLeaveSpace(ctx context.Context, userID, spaceID string) (bool, error) {
	return c.hasSpacePermission(ctx, spaceID, userID, PermServerLeave)
}

// ============================================================================
// Room-scoped permissions
// ============================================================================

// CanPostMessage checks if a user can post new root messages in a specific room.
func (c *ChattoCore) CanPostMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessagePost)
}

// CanPostInThread checks if a user can post messages in a thread.
func (c *ChattoCore) CanPostInThread(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessagePostInThread)
}

// CanReply checks if a user can use reply attribution (inReplyTo) on room-level messages.
func (c *ChattoCore) CanReply(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageReply)
}

// CanReplyInThread checks if a user can use reply attribution (inReplyTo) on thread messages.
func (c *ChattoCore) CanReplyInThread(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageReplyInThread)
}

// CanReactToMessage checks if a user can add/remove reactions in a specific room.
func (c *ChattoCore) CanReactToMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageReact)
}

// CanEchoMessage checks if a user can echo thread replies to the main channel.
func (c *ChattoCore) CanEchoMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageEcho)
}

// CanEditOwnMessage checks if a user can edit their own messages in a specific room.
func (c *ChattoCore) CanEditOwnMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageEditOwn)
}

// CanEditAnyMessage checks if a user can edit any user's messages in a specific room.
func (c *ChattoCore) CanEditAnyMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageEditAny)
}

// CanDeleteOwnMessage checks if a user can delete their own messages in a specific room.
func (c *ChattoCore) CanDeleteOwnMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageDeleteOwn)
}

// CanDeleteAnyMessage checks if a user can delete any user's messages in a specific room.
func (c *ChattoCore) CanDeleteAnyMessage(ctx context.Context, userID, spaceID, roomID string) (bool, error) {
	return c.hasRoomPermission(ctx, spaceID, roomID, userID, PermMessageDeleteAny)
}
