<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { instanceIdToSegment } from '$lib/navigation';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';

  const getInstanceId = getActiveInstance();

  // Redirect to the instance settings sub-page.
  // Uses a component instead of a load function because load functions
  // run before ensureOrigin(), so segmentToInstanceId('-') may fail.
  $effect(() => {
    goto(resolve('/chat/[instanceId]/admin/settings/instance', { instanceId: instanceIdToSegment(getInstanceId()) }), { replaceState: true });
  });
</script>
