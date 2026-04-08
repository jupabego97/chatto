<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { instanceIdToSegment } from '$lib/navigation';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';

  const getInstanceId = getActiveInstance();
  import RoomDirectory from '$lib/RoomDirectory.svelte';
  import PaneHeader from '$lib/ui/PaneHeader.svelte';
  import PageTitle from '$lib/ui/PageTitle.svelte';
  import { getSpacePermissions } from '$lib/state/space';

  const spaceId = $derived(page.params.spaceId!);

  // Get space permissions from context (set by parent layout)
  // Access .current in $derived to maintain reactivity when permissions load async
  const spacePermissions = getSpacePermissions();
  const canBrowseRooms = $derived(spacePermissions.current.canBrowseRooms);
</script>

<PageTitle title="Browse Rooms" />

{#if !canBrowseRooms}
  <div class="flex h-full w-full flex-col items-center justify-center gap-4">
    <div class="text-2xl font-semibold text-danger">Access Denied</div>
    <div class="text-lg text-muted">You do not have permission to browse rooms in this space.</div>
    <a href={resolve('/chat/[instanceId]/[spaceId]', { instanceId: instanceIdToSegment(getInstanceId()), spaceId })} class="text-primary hover:underline"
      >Return to Space</a
    >
  </div>
{:else}
  <div class="flex min-h-0 min-w-0 flex-1 flex-col">
    <PaneHeader title="Browse Rooms" showMobileNav />

    <div class="flex-1 overflow-auto p-6">
      <div class="max-w-2xl">
        <RoomDirectory {spaceId} />
      </div>
    </div>
  </div>
{/if}
