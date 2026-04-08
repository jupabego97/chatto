<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { onMount } from 'svelte';
  import { GetMyRoomsInSpaceDocument } from '$lib/gql/graphql';
  import { instanceIdToSegment } from '$lib/navigation';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { useConnection } from '$lib/state/instance/connection.svelte';
  import { getLastRoom } from '$lib/storage/lastRoom';

  let { data } = $props();

  const getInstanceId = getActiveInstance();
  const connection = useConnection();

  // Resolve and navigate to the best room in this space.
  // Runs once on mount — the parent layout remounts this component per spaceId via {#key}.
  onMount(() => {
    const { spaceId } = data;
    if (!spaceId) return;

    const instanceId = getInstanceId();

    // Check localStorage first (no network needed)
    const lastRoom = getLastRoom(instanceId, spaceId);
    if (lastRoom) {
      goto(resolve('/chat/[instanceId]/[spaceId]/[roomId]', { instanceId: instanceIdToSegment(instanceId), spaceId, roomId: lastRoom }), { replaceState: true });
      return;
    }

    // No last room — query for any room the user has joined
    connection()
      .client.query(GetMyRoomsInSpaceDocument, { spaceId })
      .toPromise()
      .then((result) => {
        const rooms = result.data?.me?.rooms ?? [];
        if (rooms.length > 0) {
          goto(resolve('/chat/[instanceId]/[spaceId]/[roomId]', { instanceId: instanceIdToSegment(instanceId), spaceId, roomId: rooms[0].id }), { replaceState: true });
        }
      });
  });
</script>

<div class="flex flex-1 items-center justify-center p-8">
  <div class="max-w-md text-center">
    <div class="mb-6">
      <span class="mb-4 iconify inline-block text-6xl text-muted uil--comments-alt"></span>
      <h2 class="mb-2 text-2xl font-bold">No Room Selected</h2>
      <p class="text-muted">
        Choose a room from your sidebar to get started. We promise this page will eventually do
        something more useful.
      </p>
    </div>
  </div>
</div>
