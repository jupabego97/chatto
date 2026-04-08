<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';
  import { getInstancePermissions } from '$lib/state/instance/permissions.svelte';
  import { resolveLastPosition } from '$lib/storage/lastRoom';

  const getInstanceId = getActiveInstance();
  const instancePerms = getInstancePermissions();

  // Navigate directly to last position, or Browse Spaces as fallback.
  // If localStorage has a position, go there immediately.
  // Otherwise, wait for permissions to decide between Browse Spaces and fallback.
  $effect(() => {
    if (sessionStorage.getItem('returnUrl')) return;

    const lastPos = resolveLastPosition(getInstanceId());
    if (lastPos) {
      // eslint-disable-next-line svelte/no-navigation-without-resolve -- lastPos from resolveLastPosition() is already resolved
      goto(lastPos, { replaceState: true });
      return;
    }

    if (!instancePerms.current.loaded) return;
    if (instancePerms.current.canListSpaces) {
      goto(resolve('/chat/spaces'), { replaceState: true });
    }
  });
</script>

{#if instancePerms.current.loaded && !instancePerms.current.canListSpaces}
  <div class="flex flex-1 items-center justify-center p-8">
    <div class="max-w-md text-center">
      <span class="mb-4 iconify inline-block text-6xl text-muted uil--comment-message"></span>
      <h2 class="mb-2 text-2xl font-bold">Welcome to Chatto!</h2>
      <p class="text-muted">Choose a space from the sidebar to get started.</p>
    </div>
  </div>
{/if}
