package core

import (
	"hmans.de/chatto/internal/core/rbac"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// Space role names (system-defined, cannot be deleted).
const (
	// SpaceRoleOwner has all space permissions and is the highest authority.
	// Must be explicitly assigned. Space creator gets this role.
	SpaceRoleOwner = "owner"

	// SpaceRoleSuspended carries explicit denies for user-behavior permissions
	// inside a space. Sits above admin so the denies can't be overridden.
	SpaceRoleSuspended = "suspended"

	// SpaceRoleAdmin has administrative permissions.
	// Must be explicitly assigned.
	SpaceRoleAdmin = "admin"

	// SpaceRoleModeration carries the heavy moderation grants (delete-any,
	// member.remove). Sits between admin and moderator in the hierarchy and
	// is granted on demand rather than baked into the moderator badge.
	SpaceRoleModeration = "moderation"

	// SpaceRoleModerator has moderation permissions.
	// Must be explicitly assigned.
	SpaceRoleModerator = "moderator"

	// SpaceRoleEveryone is implicit for all space members.
	SpaceRoleEveryone = "everyone"
)

// IsSpaceSystemRole returns true if the role name is a system role that cannot be deleted.
func IsSpaceSystemRole(name string) bool {
	return name == SpaceRoleOwner || name == SpaceRoleSuspended ||
		name == SpaceRoleAdmin || name == SpaceRoleModeration ||
		name == SpaceRoleModerator || name == SpaceRoleEveryone
}

// IsSpaceUniversalRole returns true if the role is a universal role (same name at instance and space scope).
// These roles are excluded from the "Instance Roles" section in the space admin UI because they
// already appear under Space Roles.
func IsSpaceUniversalRole(name string) bool {
	return name == SpaceRoleEveryone
}

// SpaceVirtualRoles returns the virtual role definitions for space RBAC.
// Only everyone is virtual - owner, admin, and moderator are explicitly created in KV.
// Position scheme: owner=0, suspended=1, admin=2, moderation=3, moderator=4, custom=5..n, everyone=MAX
func SpaceVirtualRoles() []*corev1.Role {
	return []*corev1.Role{
		{
			Name:        SpaceRoleEveryone,
			DisplayName: "Everyone",
			Description: "All space members",
			Position:    rbac.PositionEveryone,
		},
	}
}
