import { onSpaceEvent, onPresenceChange, type SpaceEvent } from '$lib/spaceEventBus.svelte';
import type { PresenceStatus } from '$lib/gql/graphql';

type SpaceEventHandler = (event: SpaceEvent) => void;

/**
 * Hook to subscribe to space events with automatic cleanup.
 * Must be called during component initialization (not inside conditionals).
 *
 * @example
 * useSpaceEvent((event) => {
 *   if (event.event?.__typename === 'MessagePostedEvent') {
 *     handleNewMessage(event);
 *   }
 * });
 */
export function useSpaceEvent(handler: SpaceEventHandler) {
  $effect(() => onSpaceEvent(handler));
}

/**
 * Hook to subscribe to presence change events with automatic cleanup.
 * Must be called during component initialization.
 *
 * If the space event bus is not initialized, this is a no-op.
 */
export function usePresenceChange(handler: (userId: string, status: PresenceStatus) => void) {
  $effect(() => onPresenceChange(handler));
}
