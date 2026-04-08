<script lang="ts">
  import type { Snippet } from 'svelte';
  import { sidebarNav } from '$lib/state/globals.svelte';

  let {
    children,
    width = 'md:w-64',
    mobileWidth = 'max-md:w-64'
  }: {
    children: Snippet;
    width?: string;
    mobileWidth?: string;
  } = $props();
</script>

<!--
	Secondary sidebar (room list, DM conversations, etc.)
	- Desktop: shown in normal flow with fixed width
	- Mobile: fixed overlay positioned after SpaceList
-->
<div
  class={[
    'z-50 flex min-w-0 flex-col overflow-hidden border-r border-border bg-background',
    width,
    mobileWidth,
    'md:flex-initial',
    // Mobile: fixed overlay positioned after SpaceList (~68px)
    'max-md:fixed max-md:top-11 max-md:bottom-0 max-md:left-17',
    sidebarNav.isOpen ? '' : 'hidden'
  ]}
>
  {@render children()}
</div>
