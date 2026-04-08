/**
 * Per-space and per-room notification level preferences.
 *
 * Tracks the viewer's notification level for spaces and rooms.
 * Used by SpaceList/RoomList to suppress unread indicators for muted rooms,
 * and by the preferences page to display and edit settings.
 */

import { SvelteMap } from 'svelte/reactivity';
import { NotificationLevel } from '$lib/gql/graphql';

/** Composite key for room-level preferences: "spaceId:roomId" */
function roomKey(spaceId: string, roomId: string): string {
  return `${spaceId}:${roomId}`;
}

export class NotificationLevelStore {
  /** Space-level preferences: spaceId -> { level, effectiveLevel } */
  private spaceLevels = new SvelteMap<
    string,
    { level: NotificationLevel; effectiveLevel: NotificationLevel }
  >();

  /** Room-level preferences: "spaceId:roomId" -> { level, effectiveLevel } */
  private roomLevels = new SvelteMap<
    string,
    { level: NotificationLevel; effectiveLevel: NotificationLevel }
  >();

  /**
   * Set the viewer's notification preference for a space.
   */
  setSpacePreference(
    spaceId: string,
    level: NotificationLevel,
    effectiveLevel: NotificationLevel
  ): void {
    this.spaceLevels.set(spaceId, { level, effectiveLevel });
  }

  /**
   * Set the viewer's notification preference for a room.
   */
  setRoomPreference(
    spaceId: string,
    roomId: string,
    level: NotificationLevel,
    effectiveLevel: NotificationLevel
  ): void {
    this.roomLevels.set(roomKey(spaceId, roomId), { level, effectiveLevel });
  }

  /**
   * Get the viewer's notification preference for a space.
   * Returns DEFAULT/NORMAL if not set.
   */
  getSpacePreference(spaceId: string): {
    level: NotificationLevel;
    effectiveLevel: NotificationLevel;
  } {
    return (
      this.spaceLevels.get(spaceId) ?? {
        level: NotificationLevel.Default,
        effectiveLevel: NotificationLevel.Normal
      }
    );
  }

  /**
   * Get the viewer's notification preference for a room.
   * Returns DEFAULT with the space's effective level if not set.
   */
  getRoomPreference(
    spaceId: string,
    roomId: string
  ): { level: NotificationLevel; effectiveLevel: NotificationLevel } {
    const roomPref = this.roomLevels.get(roomKey(spaceId, roomId));
    if (roomPref) return roomPref;

    // Fall back to space-level effective level
    const spacePref = this.getSpacePreference(spaceId);
    return {
      level: NotificationLevel.Default,
      effectiveLevel: spacePref.effectiveLevel
    };
  }

  /**
   * Get the effective notification level for a room.
   * Resolves: room-level -> space-level -> NORMAL.
   */
  getEffectiveLevel(spaceId: string, roomId: string): NotificationLevel {
    return this.getRoomPreference(spaceId, roomId).effectiveLevel;
  }

  /**
   * Check if a room is muted (no notifications, no unread markers).
   */
  isRoomMuted(spaceId: string, roomId: string): boolean {
    return this.getEffectiveLevel(spaceId, roomId) === NotificationLevel.Muted;
  }

  /**
   * Check if a space is fully muted (space-level muted, no room overrides).
   */
  isSpaceMuted(spaceId: string): boolean {
    return this.getSpacePreference(spaceId).effectiveLevel === NotificationLevel.Muted;
  }

  /**
   * Clear all preferences. Called on logout.
   */
  clear(): void {
    this.spaceLevels.clear();
    this.roomLevels.clear();
  }
}
