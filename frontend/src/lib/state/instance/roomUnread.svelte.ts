import { SvelteMap, SvelteSet } from 'svelte/reactivity';

/**
 * Tracks which rooms have unread messages, organized by space.
 * Space-level unread status is derived from this.
 *
 * This allows the frontend to track unread state without extra queries:
 * - NewMessageInSpaceEvent sets a room as unread
 * - Marking a room as read (posting or entering) clears it
 * - Space indicator = "any room in this space is unread"
 */
export class RoomUnreadStore {
  // Map of spaceId -> Set of roomIds with unread messages
  // Using SvelteMap for reactive method calls (.get, .set, .has, .delete, .size)
  private unreadRooms = new SvelteMap<string, SvelteSet<string>>();

  private static readonly SENTINEL = '__space_unread__';

  /**
   * Set unread status for a specific room.
   */
  setRoomUnread(spaceId: string, roomId: string, unread: boolean): void {
    if (unread) {
      let rooms = this.unreadRooms.get(spaceId);
      if (!rooms) {
        rooms = new SvelteSet<string>();
        this.unreadRooms.set(spaceId, rooms);
      }
      rooms.add(roomId);
    } else {
      const rooms = this.unreadRooms.get(spaceId);
      if (rooms) {
        rooms.delete(roomId);
        // Also clear the sentinel - we now have specific room knowledge
        rooms.delete(RoomUnreadStore.SENTINEL);
        if (rooms.size === 0) {
          this.unreadRooms.delete(spaceId);
        }
      }
    }
  }

  /**
   * Check if a space has any unread rooms.
   */
  spaceHasUnread(spaceId: string): boolean {
    const rooms = this.unreadRooms.get(spaceId);
    return rooms !== undefined && rooms.size > 0;
  }

  /**
   * Get the first known unread room ID for a space.
   * Returns null if only sentinel data exists (we don't know specific rooms).
   */
  getFirstUnreadRoomId(spaceId: string): string | null {
    const rooms = this.unreadRooms.get(spaceId);
    if (!rooms) return null;

    for (const roomId of rooms) {
      if (roomId !== RoomUnreadStore.SENTINEL) {
        return roomId;
      }
    }
    return null;
  }

  /**
   * Check if a specific room is unread.
   */
  roomIsUnread(spaceId: string, roomId: string): boolean {
    const rooms = this.unreadRooms.get(spaceId);
    return rooms !== undefined && rooms.has(roomId);
  }

  /**
   * Initialize unread state for a space from room data.
   * Call this when loading rooms for a space.
   */
  initSpaceRooms(spaceId: string, rooms: Array<{ id: string; hasUnread: boolean }>): void {
    // Clear existing state for this space
    this.unreadRooms.delete(spaceId);

    // Add unread rooms
    for (const room of rooms) {
      if (room.hasUnread) {
        this.setRoomUnread(spaceId, room.id, true);
      }
    }
  }

  /**
   * Set space-level unread status (for initial load when we only know space has unread,
   * not which specific rooms). Uses a sentinel room ID.
   *
   * WARNING: This uses a sentinel room ID. Do not call roomIsUnread() with the sentinel
   * - it's an internal implementation detail. When actual room data is loaded via
   * initSpaceRooms(), or when specific rooms are marked as read, the sentinel is cleared.
   */
  setSpaceHasUnread(spaceId: string, hasUnread: boolean): void {
    if (hasUnread) {
      // Use a sentinel to indicate "space has unread but we don't know which rooms"
      this.setRoomUnread(spaceId, RoomUnreadStore.SENTINEL, true);
    } else {
      // Clear all rooms for this space
      this.unreadRooms.delete(spaceId);
    }
  }

  /**
   * Get count of spaces with unread rooms (for PWA badge).
   */
  get unreadSpaceCount(): number {
    return this.unreadRooms.size;
  }

  /**
   * Whether any space has unread rooms (for PWA flag badge).
   */
  get hasAnyUnread(): boolean {
    return this.unreadRooms.size > 0;
  }

  /**
   * Clear all unread state.
   */
  clear(): void {
    this.unreadRooms.clear();
  }
}
