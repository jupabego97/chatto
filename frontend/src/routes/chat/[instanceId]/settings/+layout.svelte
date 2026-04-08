<script lang="ts">
  import { resolve } from '$app/paths';
  import { getCurrentUser } from '$lib/auth/currentUser.svelte';
  import { instanceIdToSegment } from '$lib/navigation';
  import { getActiveInstance } from '$lib/state/activeInstance.svelte';

  const getInstanceId = getActiveInstance();
  const instanceSegment = $derived(instanceIdToSegment(getInstanceId()));
  import SecondarySidebar from '$lib/components/SecondarySidebar.svelte';
  import SidebarNav from '$lib/components/SidebarNav.svelte';
  import LoadingPage from '$lib/ui/LoadingPage.svelte';

  let { children } = $props();

  const currentUser = getCurrentUser();

  // Nav items for settings
  const navItems = $derived([
    { href: resolve('/chat/[instanceId]/settings', { instanceId: instanceSegment }), label: 'Profile', icon: 'iconify uil--user' },
    { href: resolve('/chat/[instanceId]/settings/preferences', { instanceId: instanceSegment }), label: 'Preferences', icon: 'iconify uil--clock' },
    { href: resolve('/chat/[instanceId]/settings/account', { instanceId: instanceSegment }), label: 'Account', icon: 'iconify uil--setting' },
    { href: resolve('/chat/[instanceId]/settings/notifications', { instanceId: instanceSegment }), label: 'Notifications', icon: 'iconify uil--bell' }
  ]);
</script>

{#if currentUser.loading}
  <LoadingPage />
{:else if !currentUser.user}
  <LoadingPage message="Not logged in" />
{:else}
  <SecondarySidebar width="md:w-56">
    <SidebarNav title="Settings" items={navItems} backHref={resolve('/chat/[instanceId]', { instanceId: instanceSegment })} />
  </SecondarySidebar>

  <!-- Main content -->
  <div class="flex min-h-0 min-w-0 flex-1 flex-col">
    {@render children?.()}
  </div>
{/if}
