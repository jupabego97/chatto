package rbac

import (
	"errors"
	"regexp"
	"strings"
)

// Validation errors.
var (
	// ErrInvalidRoleName is returned when a role name doesn't match the required format.
	ErrInvalidRoleName = errors.New("invalid role name: space roles must be lowercase letters only (a-z), 1-32 chars")

	// ErrInvalidInstanceRoleName is returned when an instance role name doesn't match the required format.
	ErrInvalidInstanceRoleName = errors.New("invalid instance role name: must start with 'instance-' followed by lowercase letters only, max 32 chars total")

	// ErrReservedRoleName is returned when attempting to use a reserved role name.
	ErrReservedRoleName = errors.New("role name 'instance' is reserved and cannot be used for space roles")

	// ErrRoleNotFound is returned when a role doesn't exist.
	ErrRoleNotFound = errors.New("role not found")

	// ErrRoleAlreadyExists is returned when attempting to create a role that already exists.
	ErrRoleAlreadyExists = errors.New("role already exists")

	// ErrCannotDeleteSystemRole is returned when attempting to delete a system role.
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")

	// ErrInvalidPermission is returned when a permission value is not recognized.
	ErrInvalidPermission = errors.New("invalid permission")

	// ErrPermissionDenied is returned when a user lacks the required permission.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrCannotAssignHigherRole is returned when a user tries to assign a role equal to or higher than their own.
	ErrCannotAssignHigherRole = errors.New("cannot assign role equal to or higher than your own")

	// ErrCannotRevokeHigherRole is returned when a user tries to revoke a role equal to or higher than their own.
	ErrCannotRevokeHigherRole = errors.New("cannot revoke role equal to or higher than your own")

	// ErrCannotManageHigherUser is returned when a user tries to manage another user with equal or higher role.
	ErrCannotManageHigherUser = errors.New("cannot manage user with equal or higher role")

	// ErrCannotReorderSystemRole is returned when attempting to reorder a system role.
	ErrCannotReorderSystemRole = errors.New("cannot reorder system roles")
)

// spaceRoleNameRegex matches valid space role names: lowercase letters only, 1-32 characters.
// Space roles must be single lowercase words (no numbers, no dashes).
var spaceRoleNameRegex = regexp.MustCompile(`^[a-z]{1,32}$`)

// instanceRoleNameRegex matches valid instance role names: must start with "instance-"
// followed by lowercase letters only, total max 32 characters.
var instanceRoleNameRegex = regexp.MustCompile(`^instance-[a-z]{1,23}$`)

// ValidateSpaceRoleName checks if a space role name is valid.
// Valid names: lowercase letters only (a-z), 1-32 characters.
// The name "instance" is reserved and cannot be used.
func ValidateSpaceRoleName(name string) error {
	if name == "instance" {
		return ErrReservedRoleName
	}
	if !spaceRoleNameRegex.MatchString(name) {
		return ErrInvalidRoleName
	}
	return nil
}

// ValidateInstanceRoleName checks if an instance role name is valid.
// Valid names: must start with "instance-" followed by lowercase letters only,
// max 32 characters total (so the suffix can be up to 23 characters).
func ValidateInstanceRoleName(name string) error {
	if !instanceRoleNameRegex.MatchString(name) {
		return ErrInvalidInstanceRoleName
	}
	return nil
}

// ValidateRoleName checks if a role name is valid based on its type.
// For space roles (no "instance-" prefix): lowercase letters only, 1-32 chars, not "instance".
// For instance roles (with "instance-" prefix): must be well-formed.
//
// This function auto-detects the role type based on the prefix.
func ValidateRoleName(name string) error {
	if strings.HasPrefix(name, "instance-") {
		return ValidateInstanceRoleName(name)
	}
	return ValidateSpaceRoleName(name)
}
