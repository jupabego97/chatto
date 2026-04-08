package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/nats-io/nats.go/jetstream"

	"hmans.de/chatto/internal/core/rbac"
)

// PermissionResolver handles permission resolution using role hierarchy:
//
// At all levels (instance, space, room), the role with the highest rank
// (lowest position number) whose explicit grant/deny decision is found first wins.
//
// Resolution rules:
// 1. Get user's roles sorted by hierarchy (lower position = higher rank)
// 2. For each role, check for explicit grant or deny
// 3. First explicit decision found → that's the answer
// 4. If no explicit decision at current level → fall back to parent level
//
// This enables patterns like:
// - #announcements rooms where "everyone" is denied message.post but
//   "owner/admin/moderator" can still post because they have higher rank
// - Instance admin not being blocked by a "everyone" denial because
//   admin is checked first in the hierarchy
type PermissionResolver struct {
	core *ChattoCore
}

// NewPermissionResolver creates a new permission resolver.
func NewPermissionResolver(core *ChattoCore) *PermissionResolver {
	return &PermissionResolver{core: core}
}

// HasInstancePermission checks if a user has a permission at the instance level.
// Only checks instance-level roles and KV. Used for permissions that only apply
// at instance scope (like admin.access, space.create, dm.view).
func (r *PermissionResolver) HasInstancePermission(ctx context.Context, userID string, perm Permission) (bool, error) {
	// Validate permission applies at instance scope (if it's a known permission)
	if meta, known := GetPermissionMetadata(perm); known && !permissionMetadataHasScope(meta, ScopeInstance) {
		return false, fmt.Errorf("permission %s does not apply at instance scope", perm)
	}

	return r.resolveInstancePermission(ctx, userID, string(perm))
}

// permissionMetadataHasScope checks if a permission applies at the given scope.
func permissionMetadataHasScope(meta PermissionMetadata, scope PermissionScope) bool {
	for _, s := range meta.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasSpacePermission checks if a user has a permission at the space level.
//
// Uses the deny-always-wins model: all denials across all levels are checked
// first, then grants are checked in authority order (instance → space).
//
// For space-scoped permissions (like room.create), the user must be a space
// member. space.join is exempt (non-members need it to join).
func (r *PermissionResolver) HasSpacePermission(ctx context.Context, userID, spaceID string, perm Permission) (bool, error) {
	// Validate permission applies at space scope (if it's a known permission)
	if meta, known := GetPermissionMetadata(perm); known {
		if !permissionMetadataHasScope(meta, ScopeSpace) && !permissionMetadataHasScope(meta, ScopeInstance) {
			return false, fmt.Errorf("permission %s does not apply at space scope", perm)
		}
	}

	// DM space uses simplified permissions
	if IsDMSpace(spaceID) {
		return r.resolveDMPermission(perm), nil
	}

	parts := perm.KeyParts()
	if parts.Verb == "" || parts.ObjectType == "" {
		return false, nil
	}

	// Membership gate: space-scoped permissions require membership
	// (except space.join and space.list, which non-members need for discovery)
	if PermissionAppliesAtScope(perm, ScopeSpace) && perm != PermSpaceJoin && perm != PermSpaceList {
		isMember, err := r.core.SpaceMembershipExists(ctx, userID, spaceID)
		if err != nil {
			return false, fmt.Errorf("failed to check space membership: %w", err)
		}
		if !isMember {
			return false, nil
		}
	}

	// Gather roles
	instanceRoles, err := r.getUserInstanceRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	spaceRoles, err := r.getUserSpaceRoles(ctx, spaceID, userID)
	if err != nil {
		return false, err
	}

	// Universal roles (everyone) appear in both lists.
	// Filter them from instanceRoles when checking space KV to avoid redundant lookups.
	instanceOnlyRoles := filterOutSpaceRoles(instanceRoles, spaceRoles)

	instanceKV := r.core.instanceRBACEngine.KV()
	spaceKV, err := r.core.getSpaceRBACKV(ctx, spaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get space RBAC KV: %w", err)
	}

	// === PHASE 1: Check ALL denials across all levels ===

	// Instance-level denials (instance roles in instance KV)
	for _, role := range instanceRoles {
		denied, err := r.keyExists(ctx, instanceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by instance role", "role", role, "permission", string(perm), "user", userID)
			return false, nil
		}
	}

	// Space-level denials: space roles in space KV
	for _, role := range spaceRoles {
		denied, err := r.keyExists(ctx, spaceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by space role", "role", role, "permission", string(perm), "space", spaceID, "user", userID)
			return false, nil
		}
	}

	// Space-level denials: instance-only roles in space KV (overrides)
	for _, role := range instanceOnlyRoles {
		denied, err := r.keyExists(ctx, spaceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by instance role (space config)", "role", role, "permission", string(perm), "space", spaceID, "user", userID)
			return false, nil
		}
	}

	// === PHASE 2: Check grants in authority order (instance → space) ===

	// Instance grants (instance roles in instance KV)
	if PermissionAppliesAtScope(perm, ScopeInstance) {
		for _, role := range instanceRoles {
			granted, err := r.keyExists(ctx, instanceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
			if err != nil {
				return false, err
			}
			if granted {
				return true, nil
			}
		}
	}

	// Space grants: space roles in space KV
	for _, role := range spaceRoles {
		granted, err := r.keyExists(ctx, spaceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if granted {
			return true, nil
		}
	}

	// Space grants: instance-only roles in space KV (overrides)
	for _, role := range instanceOnlyRoles {
		granted, err := r.keyExists(ctx, spaceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if granted {
			return true, nil
		}
	}

	return false, nil
}

// HasRoomPermission checks if a user has a permission at the room level.
//
// Resolution order:
// 1. Instance/space denials (deny-always-wins at these levels)
// 2. Room-level permissions: grants checked BEFORE denials
//   - Room grants allow per-room overrides (e.g., owner can post in #announcements
//     even though "everyone" is denied)
// 3. Instance/space grants (fallback when no room-level decision)
func (r *PermissionResolver) HasRoomPermission(ctx context.Context, userID, spaceID, roomID string, perm Permission) (bool, error) {
	// Validate permission applies at room scope
	if !PermissionAppliesAtScope(perm, ScopeRoom) && !PermissionAppliesAtScope(perm, ScopeSpace) && !PermissionAppliesAtScope(perm, ScopeInstance) {
		return false, fmt.Errorf("permission %s does not apply at room scope", perm)
	}

	// DM space uses simplified permissions
	if IsDMSpace(spaceID) {
		return r.resolveDMPermission(perm), nil
	}

	parts := perm.KeyParts()
	if parts.Verb == "" || parts.ObjectType == "" {
		return false, nil
	}

	// Gather roles
	instanceRoles, err := r.getUserInstanceRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	spaceRoles, err := r.getUserSpaceRoles(ctx, spaceID, userID)
	if err != nil {
		return false, err
	}

	// Universal roles (everyone) appear in both lists.
	// Filter them from instanceRoles when checking space KV to avoid redundant lookups.
	instanceOnlyRoles := filterOutSpaceRoles(instanceRoles, spaceRoles)

	instanceKV := r.core.instanceRBACEngine.KV()
	spaceKV, err := r.core.getSpaceRBACKV(ctx, spaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get space RBAC KV: %w", err)
	}

	// === PHASE 1: Check ALL denials across ALL levels ===

	// Instance-level denials (instance roles in instance KV)
	for _, role := range instanceRoles {
		denied, err := r.keyExists(ctx, instanceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by instance role", "role", role, "permission", string(perm), "room", roomID, "user", userID)
			return false, nil
		}
	}

	// Space-level denials: space roles in space KV (objectId=any)
	for _, role := range spaceRoles {
		denied, err := r.keyExists(ctx, spaceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by space role", "role", role, "permission", string(perm), "room", roomID, "user", userID)
			return false, nil
		}
	}

	// Space-level denials: instance-only roles in space KV (objectId=any)
	for _, role := range instanceOnlyRoles {
		denied, err := r.keyExists(ctx, spaceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by instance role (space config)", "role", role, "permission", string(perm), "room", roomID, "user", userID)
			return false, nil
		}
	}

	// Room-level permissions: Use role hierarchy (lower position = higher rank wins).
	// For each role in hierarchy order, check for explicit grant or deny. First found wins.
	// This enables patterns like #announcements where "owner" grant beats "everyone" deny.
	if PermissionAppliesAtScope(perm, ScopeRoom) {
		// Get space roles sorted by hierarchy (lower position = higher rank = checked first)
		rolesWithPos, err := r.getUserSpaceRolesWithPositions(ctx, spaceID, userID)
		if err != nil {
			return false, err
		}

		// Check each role in hierarchy order - first explicit permission wins
		for _, rp := range rolesWithPos {
			// Check for room-level grant
			granted, err := r.keyExists(ctx, spaceKV, rbac.AllowKey(rp.name, parts.Verb, parts.ObjectType, roomID))
			if err != nil {
				return false, err
			}
			if granted {
				r.core.logger.Debug("Permission granted by space role (room config, hierarchy)", "role", rp.name, "position", rp.position, "permission", string(perm), "room", roomID, "user", userID)
				return true, nil
			}

			// Check for room-level deny
			denied, err := r.keyExists(ctx, spaceKV, rbac.DenyKey(rp.name, parts.Verb, parts.ObjectType, roomID))
			if err != nil {
				return false, err
			}
			if denied {
				r.core.logger.Debug("Permission denied by space role (room config, hierarchy)", "role", rp.name, "position", rp.position, "permission", string(perm), "room", roomID, "user", userID)
				return false, nil
			}
		}

		// Also check instance-only roles (not part of space hierarchy, checked after space roles)
		for _, role := range instanceOnlyRoles {
			granted, err := r.keyExists(ctx, spaceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, roomID))
			if err != nil {
				return false, err
			}
			if granted {
				r.core.logger.Debug("Permission granted by instance role (room config)", "role", role, "permission", string(perm), "room", roomID, "user", userID)
				return true, nil
			}

			denied, err := r.keyExists(ctx, spaceKV, rbac.DenyKey(role, parts.Verb, parts.ObjectType, roomID))
			if err != nil {
				return false, err
			}
			if denied {
				r.core.logger.Debug("Permission denied by instance role (room config)", "role", role, "permission", string(perm), "room", roomID, "user", userID)
				return false, nil
			}
		}
	}

	// === PHASE 2: Check instance and space grants (fallback when no room-level decision) ===

	// Instance grants (instance roles in instance KV)
	if PermissionAppliesAtScope(perm, ScopeInstance) {
		for _, role := range instanceRoles {
			granted, err := r.keyExists(ctx, instanceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
			if err != nil {
				return false, err
			}
			if granted {
				return true, nil
			}
		}
	}

	// Space grants: space roles in space KV (objectId=any)
	if PermissionAppliesAtScope(perm, ScopeSpace) {
		for _, role := range spaceRoles {
			granted, err := r.keyExists(ctx, spaceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
			if err != nil {
				return false, err
			}
			if granted {
				return true, nil
			}
		}

		// Space grants: instance-only roles in space KV (objectId=any)
		for _, role := range instanceOnlyRoles {
			granted, err := r.keyExists(ctx, spaceKV, rbac.AllowKey(role, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
			if err != nil {
				return false, err
			}
			if granted {
				return true, nil
			}
		}
	}

	return false, nil
}

// ============================================================================
// Internal Resolution Methods
// ============================================================================

// resolveInstancePermission resolves a permission at the instance level only.
// Used by HasInstancePermission which only checks instance KV.
// Uses role hierarchy: higher-ranked role's explicit decision wins.
func (r *PermissionResolver) resolveInstancePermission(ctx context.Context, userID, perm string) (bool, error) {
	engine := r.core.instanceRBACEngine

	parts := Permission(perm).KeyParts()
	if parts.Verb == "" || parts.ObjectType == "" {
		return false, nil
	}

	// Get roles sorted by hierarchy (lower position = higher rank = checked first)
	rolesWithPos, err := r.getUserInstanceRolesWithPositions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check each role in hierarchy order - first explicit permission wins
	for _, rp := range rolesWithPos {
		// Check for grant
		granted, err := r.keyExists(ctx, engine.KV(), rbac.AllowKey(rp.name, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if granted {
			r.core.logger.Debug("Permission granted by instance role (hierarchy)", "role", rp.name, "position", rp.position, "permission", perm, "user", userID)
			return true, nil
		}

		// Check for deny
		denied, err := r.keyExists(ctx, engine.KV(), rbac.DenyKey(rp.name, parts.Verb, parts.ObjectType, rbac.ObjectIdAny))
		if err != nil {
			return false, err
		}
		if denied {
			r.core.logger.Debug("Permission denied by instance role (hierarchy)", "role", rp.name, "position", rp.position, "permission", perm, "user", userID)
			return false, nil
		}
	}

	return false, nil
}

// resolveDMPermission returns whether a permission is allowed in DM context.
// DM space uses simplified permissions - only certain actions are allowed.
func (r *PermissionResolver) resolveDMPermission(perm Permission) bool {
	switch perm {
	case PermMessagePost, PermMessageEditOwn, PermMessageDeleteOwn, PermMessageReact,
		PermMessageReply, PermRoomJoin, PermRoomLeave:
		return true
	default:
		return false
	}
}

// ============================================================================
// Helper Methods
// ============================================================================

// keyExists checks if a key exists in a KV bucket.
func (r *PermissionResolver) keyExists(ctx context.Context, kv jetstream.KeyValue, key string) (bool, error) {
	_, err := kv.Get(ctx, key)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check key %s: %w", key, err)
}

// getUserInstanceRoles returns the user's instance roles (including implicit ones).
func (r *PermissionResolver) getUserInstanceRoles(ctx context.Context, userID string) ([]string, error) {
	roles, err := r.core.GetUserInstanceRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user instance roles: %w", err)
	}

	// Always include "everyone" for authenticated users
	if !slices.Contains(roles, InstRoleEveryone) {
		roles = append(roles, InstRoleEveryone)
	}

	return roles, nil
}

// filterOutSpaceRoles returns instance roles that don't appear in spaceRoles.
// Universal roles (everyone) appear in both lists; this avoids redundant
// space KV lookups since they're already checked via the space roles loop.
func filterOutSpaceRoles(instanceRoles, spaceRoles []string) []string {
	spaceSet := make(map[string]struct{}, len(spaceRoles))
	for _, r := range spaceRoles {
		spaceSet[r] = struct{}{}
	}
	var result []string
	for _, r := range instanceRoles {
		if _, ok := spaceSet[r]; !ok {
			result = append(result, r)
		}
	}
	return result
}

// getUserSpaceRoles returns the user's space roles (including implicit everyone role if member).
func (r *PermissionResolver) getUserSpaceRoles(ctx context.Context, spaceID, userID string) ([]string, error) {
	roles, err := r.core.GetUserRoles(ctx, spaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user space roles: %w", err)
	}

	return roles, nil
}

// roleWithPosition pairs a role name with its position for hierarchy sorting.
type roleWithPosition struct {
	name     string
	position int32
}

// getUserSpaceRolesWithPositions returns user's space roles with positions, sorted by hierarchy.
// Lower position = higher rank (checked first in permission resolution).
func (r *PermissionResolver) getUserSpaceRolesWithPositions(ctx context.Context, spaceID, userID string) ([]roleWithPosition, error) {
	roleNames, err := r.getUserSpaceRoles(ctx, spaceID, userID)
	if err != nil {
		return nil, err
	}

	engine, err := r.core.spaceRBACEngine(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space RBAC engine: %w", err)
	}

	result := make([]roleWithPosition, 0, len(roleNames))
	for _, name := range roleNames {
		pos := rbac.PositionEveryone // Default for virtual roles or if lookup fails
		if role, err := engine.GetRole(ctx, name); err == nil && role != nil {
			pos = role.Position
		}
		result = append(result, roleWithPosition{name: name, position: pos})
	}

	// Sort by position ascending (lower = higher rank = checked first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].position < result[j].position
	})

	return result, nil
}

// getUserInstanceRolesWithPositions returns user's instance roles with positions, sorted by hierarchy.
// Lower position = higher rank (checked first in permission resolution).
func (r *PermissionResolver) getUserInstanceRolesWithPositions(ctx context.Context, userID string) ([]roleWithPosition, error) {
	roleNames, err := r.getUserInstanceRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	engine := r.core.instanceRBACEngine

	result := make([]roleWithPosition, 0, len(roleNames))
	for _, name := range roleNames {
		pos := rbac.PositionEveryone // Default for virtual roles or if lookup fails
		if role, err := engine.GetRole(ctx, name); err == nil && role != nil {
			pos = role.Position
		}
		result = append(result, roleWithPosition{name: name, position: pos})
	}

	// Sort by position ascending (lower = higher rank = checked first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].position < result[j].position
	})

	return result, nil
}
