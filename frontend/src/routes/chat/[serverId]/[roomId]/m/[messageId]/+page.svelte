<!--
  Message link resolver. Fetches the event and redirects to the correct
  room (or thread) URL, with the highlight intent delivered via
  PendingHighlightStore so the destination URL stays clean (refresh won't
  re-fire the highlight). Renders nothing — the goto() fires on mount.
-->
<script lang="ts" module>
  import { graphql } from '$lib/gql';
  import { goto } from '$app/navigation';
  import {
    roomPathForSegment,
    roomThreadPathForSegment,
    type RoomRouteKind
  } from '$lib/roomUrls';
  import type { Client } from '@urql/svelte';
  import type { PendingHighlightStore } from '$lib/state/server/pendingHighlight.svelte';

  const ResolveMessageLinkQuery = graphql(`
    query ResolveMessageLink($roomId: ID!, $eventId: ID!) {
      room(roomId: $roomId) {
        event(eventId: $eventId) {
          id
          event {
            __typename
            ... on MessagePostedEvent {
              threadRootEventId
            }
          }
        }
      }
    }
  `);

  /**
   * Fetch a message by ID and redirect to the appropriate room or thread URL.
   * If the message is a thread reply, opens the thread pane. If not found or
   * on error, falls back to the room URL.
   */
  export async function resolveAndRedirect(
    client: Client,
    pendingHighlights: PendingHighlightStore,
    serverSegment: string,
    roomId: string,
    canonicalRoomSegment: string,
    canonicalRouteKind: RoomRouteKind,
    messageId: string
  ): Promise<void> {
    try {
      const result = await client
        .query(ResolveMessageLinkQuery, { roomId, eventId: messageId }, { requestPolicy: 'network-only' })
        .toPromise();

      const event = result.data?.room?.event;
      if (!event) {
        pendingHighlights.set(roomId, null, messageId);
        goto(roomPathForSegment(serverSegment, canonicalRoomSegment, canonicalRouteKind), { replaceState: true });
        return;
      }

      const inner = event.event;
      const threadRoot =
        inner?.__typename === 'MessagePostedEvent' ? inner.threadRootEventId : null;

      if (threadRoot) {
        pendingHighlights.set(roomId, threadRoot, messageId);
        goto(
          roomThreadPathForSegment(serverSegment, canonicalRoomSegment, threadRoot, canonicalRouteKind),
          { replaceState: true }
        );
        return;
      }

      pendingHighlights.set(roomId, null, messageId);
      goto(roomPathForSegment(serverSegment, canonicalRoomSegment, canonicalRouteKind), { replaceState: true });
    } catch {
      goto(roomPathForSegment(serverSegment, canonicalRoomSegment, canonicalRouteKind), { replaceState: true });
    }
  }
</script>

<script lang="ts">
  import { page } from '$app/state';
  import { useConnection } from '$lib/state/server/connection.svelte';
  import { serverRegistry } from '$lib/state/server/registry.svelte';
  import { getActiveServer } from '$lib/state/activeServer.svelte';

  const connection = useConnection();
  const stores = $derived(serverRegistry.getStore(getActiveServer()));
  let resolveRequest = 0;

  // Wait for the active server's rooms store to settle before redirecting,
  // so a deep-link to a DM doesn't briefly resolve as a missing channel
  // room and trigger the not-found redirect.
  const roomsStore = $derived(stores.rooms);

  $effect(() => {
    if (roomsStore.isInitialLoading) return;
    const request = ++resolveRequest;
    const routeKind = page.route?.id?.includes('/r/[roomId]') ? 'name' : 'legacy-id';
    stores.roomRoutes.resolve(page.params.roomId!, routeKind).then((resolved) => {
      if (request !== resolveRequest || !resolved) return;
      resolveAndRedirect(
        connection().client,
        stores.pendingHighlights,
        page.params.serverId!,
        resolved.roomId,
        resolved.canonicalSegment,
        resolved.canonicalRouteKind,
        page.params.messageId!
      );
    });
  });
</script>
