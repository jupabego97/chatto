import { createContext } from 'svelte';

export type RoomPermissions = {
  canPostMessage: boolean;
  canPostInThread: boolean;
  canReply: boolean;
  canReplyInThread: boolean;
  canReact: boolean;
  canEditOwnMessage: boolean;
  canEditAnyMessage: boolean;
  canDeleteOwnMessage: boolean;
  canDeleteAnyMessage: boolean;
};

export const DEFAULT_ROOM_PERMISSIONS: RoomPermissions = {
  canPostMessage: false,
  canPostInThread: false,
  canReply: false,
  canReplyInThread: false,
  canReact: false,
  canEditOwnMessage: false,
  canEditAnyMessage: false,
  canDeleteOwnMessage: false,
  canDeleteAnyMessage: false
};

const [getRoomPermissionsState, setRoomPermissionsState] = createContext<{
  current: RoomPermissions;
}>();

/**
 * Creates and sets the room permissions context.
 * Accepts a getter that computes permissions reactively — no $effect needed.
 * Must be called synchronously during component initialization.
 */
export function createRoomPermissions(getPermissions: () => RoomPermissions): void {
  setRoomPermissionsState({
    get current() {
      return getPermissions();
    }
  });
}

/**
 * Gets the current room permissions from context.
 */
export function getRoomPermissions(): RoomPermissions {
  const state = getRoomPermissionsState();
  return state.current;
}
