<script lang="ts">
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { browser } from '$app/environment';
  import { setCurrentUser } from '$lib/auth/currentUser.svelte';
  import { DM_SPACE_ID } from '$lib/constants';
  import { segmentToInstanceId } from '$lib/navigation';
  import { setActiveInstance } from '$lib/state/activeInstance.svelte';
  import { instanceRegistry } from '$lib/state/instance/registry.svelte';
  import { graphqlClientManager } from '$lib/state/instance/graphqlClient.svelte';
  import { provideConnection } from '$lib/state/instance/connection.svelte';
  import { provideInstanceEventBus } from '$lib/instanceEventBus.svelte';
  import SpaceEventProvider from '$lib/components/chat/SpaceEventProvider.svelte';

  let { children } = $props();

  const instanceId = $derived(
    segmentToInstanceId(page.params.instanceSegment ?? '-') ?? instanceRegistry.originInstance?.id ?? ''
  );

  setActiveInstance(() => instanceId);
  provideConnection(() => graphqlClientManager.getClient(instanceId));

  // eslint-disable-next-line svelte/no-unused-svelte-ignore -- Svelte compiler warning, not ESLint
  // svelte-ignore state_referenced_locally - instanceId is stable per component lifetime
  const store = instanceId ? instanceRegistry.getStore(instanceId) : undefined;
  const currentUserState = store?.currentUser;
  if (currentUserState) {
    setCurrentUser(currentUserState);
  }

  // Provide this instance's event bus to child components via Svelte context.
  // eslint-disable-next-line svelte/no-unused-svelte-ignore -- Svelte compiler warning, not ESLint
  // svelte-ignore state_referenced_locally - instanceId is stable per component lifetime
  provideInstanceEventBus(instanceId);

  // Auth guard: redirect unauthenticated users to landing page.
  $effect(() => {
    if (!browser || !currentUserState) return;
    if (currentUserState.loading) return;
    if (currentUserState.user) return;

    const currentUrl = page.url.pathname + page.url.search;
    sessionStorage.setItem('returnUrl', currentUrl);
    goto(resolve('/'), { replaceState: true });
  });
</script>

{#if currentUserState?.user}
  {#key instanceId}
    <SpaceEventProvider spaceId={DM_SPACE_ID}>
      {@render children?.()}
    </SpaceEventProvider>
  {/key}
{/if}
