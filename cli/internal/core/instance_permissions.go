package core

import (
	"hmans.de/chatto/internal/core/rbac"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// Instance role names (system-defined, cannot be deleted).
// Explicit instance roles have the "instance-" prefix.
const (
	// InstRoleOwner has all instance permissions and is the highest authority.
	// Must be explicitly assigned to users. First user gets this role.
	InstRoleOwner = "instance-owner"

	// InstRoleSuspended carries explicit denies for user-behavior permissions.
	// Sits between owner and admin in the hierarchy so its denies can't be
	// overridden by any lower-ranked role's grants.
	InstRoleSuspended = "instance-suspended"

	// InstRoleAdmin has all instance permissions.
	// Must be explicitly assigned to users.
	InstRoleAdmin = "instance-admin"

	// InstRoleModerator has moderation permissions (no admin.* permissions).
	// Must be explicitly assigned to users.
	InstRoleModerator = "instance-moderator"

	// InstRoleEveryone is implicit for all authenticated users.
	// Universal role — same name at instance and space scope.
	InstRoleEveryone = "everyone"
)

// IsInstanceSystemRole returns true if the role name is a system role.
func IsInstanceSystemRole(name string) bool {
	return name == InstRoleOwner || name == InstRoleSuspended ||
		name == InstRoleAdmin || name == InstRoleModerator ||
		name == InstRoleEveryone
}

// InstanceVirtualRoles returns the virtual role definitions for instance RBAC.
// Only everyone is virtual - owner, admin, moderator are explicitly created in KV.
// Position scheme: owner=0, suspended=1, admin=2, moderator=4, custom=5..n, everyone=MAX
func InstanceVirtualRoles() []*corev1.Role {
	return []*corev1.Role{
		{
			Name:        InstRoleEveryone,
			DisplayName: "Everyone",
			Description: "All authenticated users",
			Position:    rbac.PositionEveryone,
		},
	}
}
