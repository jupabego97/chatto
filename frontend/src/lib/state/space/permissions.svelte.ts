import { createContext } from 'svelte';

export type SpacePermissions = {
  hasAnyAdminPermission: boolean;
  canManage: boolean;
  canBrowseRooms: boolean;
  canManageRooms: boolean;
  canManageRoles: boolean;
  canAssignRoles: boolean;
  canInviteMembers: boolean;
};

const [getSpacePermissionsState, setSpacePermissionsState] = createContext<{
  current: SpacePermissions;
}>();

/**
 * Creates and sets the space permissions context.
 * Must be called synchronously during component initialization.
 * Returns a function to update the permissions.
 */
export function createSpacePermissions(): (permissions: SpacePermissions) => void {
  const state = $state<{ current: SpacePermissions }>({
    current: {
      hasAnyAdminPermission: false,
      canManage: false,
      canBrowseRooms: false,
      canManageRooms: false,
      canManageRoles: false,
      canAssignRoles: false,
      canInviteMembers: false
    }
  });
  setSpacePermissionsState(state);

  return (permissions: SpacePermissions) => {
    state.current = permissions;
  };
}

/**
 * Gets the reactive space permissions state from context.
 * Returns the wrapper object so consumers can access `.current` reactively.
 */
export function getSpacePermissions(): { current: SpacePermissions } {
  return getSpacePermissionsState();
}
