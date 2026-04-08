<script lang="ts" module>
  /* eslint-disable svelte/no-navigation-without-resolve -- href is a prop; callers pass already-resolved paths */
  // Request 96x96 for 2x retina (displayed at 48x48 CSS pixels)
  export const SpaceIconFragment = graphql(`
    fragment SpaceIconSpace on Space {
      id
      name
      logoUrl(width: 96, height: 96)
    }
  `);
</script>

<script lang="ts">
  import SpaceLogo from './components/SpaceLogo.svelte';
  import { graphql } from './gql';
  import type { SpaceIconSpaceFragment } from './gql/graphql';

  let {
    space,
    icon,
    href,
    selected = false,
    hasNotification = false,
    hasUnread = false,
    onNotificationClick,
    onUnreadClick,
    title
  }: {
    space?: SpaceIconSpaceFragment;
    /** Icon class name for icon-only mode (e.g., "iconify uil--comment-alt-lines") */
    icon?: string;
    href: string;
    selected?: boolean;
    hasNotification?: boolean;
    hasUnread?: boolean;
    onNotificationClick?: (event: MouseEvent) => void;
    onUnreadClick?: (event: MouseEvent) => void;
    title?: string;
  } = $props();
</script>

<div class="relative">
  <a
    {href}
    {title}
    aria-label={title ?? space?.name}
    class={['space-icon space-list-item cursor-pointer', selected && 'space-list-item-active']}
    data-testid={space ? 'space-icon' : icon ? 'nav-icon' : undefined}
  >
    {#if space}
      <SpaceLogo {space} />
    {:else if icon}
      <span class={icon}></span>
    {/if}
  </a>

  <!-- Notification badge -->
  {#if hasNotification}
    {#if onNotificationClick}
      <!-- Clickable notification dot - navigates to notification source -->
      <button
        type="button"
        onclick={(e) => {
          e.stopPropagation();
          onNotificationClick(e);
        }}
        class="absolute -top-1.5 -right-1.5 z-10 flex h-6 w-6 cursor-pointer items-center justify-center notification-dot"
        aria-label="Go to notification"
      >
        <span class="h-3 w-3 rounded-full bg-warning shadow-sm ring-2 ring-background"></span>
      </button>
    {:else}
      <!-- Non-clickable notification dot (backward compatible) -->
      <span
        class="absolute top-0 right-0 z-10 h-3 w-3 rounded-full bg-warning shadow-sm ring-2 ring-background"
        aria-hidden="true"
      ></span>
    {/if}

    <!-- Unread badge -->
  {:else if hasUnread}
    {#if onUnreadClick}
      <!-- Clickable unread dot - navigates to first unread room -->
      <button
        type="button"
        onclick={(e) => {
          e.stopPropagation();
          onUnreadClick(e);
        }}
        class="absolute -top-1.5 -right-1.5 z-10 flex h-6 w-6 cursor-pointer items-center justify-center notification-dot"
        aria-label="Go to first unread room"
      >
        <span
          class="h-3 w-3 rounded-full bg-muted shadow-sm ring-2 ring-background"
          data-testid="space-unread-dot"
        ></span>
      </button>
    {:else}
      <!-- Non-clickable unread dot (backward compatible) -->
      <span
        class="absolute top-0 right-0 z-10 h-3 w-3 rounded-full bg-muted shadow-sm ring-2 ring-background"
        aria-hidden="true"
        data-testid="space-unread-dot"
      ></span>
    {/if}
  {/if}
</div>
