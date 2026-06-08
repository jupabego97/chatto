<!--
@component

Shows a user's profile card. On desktop, renders as a floating popover anchored to the trigger
element. On mobile (touch devices), renders as a bottom sheet. This dual behavior comes from
ContextMenu, which handles both modes automatically.

**Props:**
- `user` - The user to display (must include id, login, displayName, presenceStatus)
- `anchorRect` - Bounding rect of the trigger element (used for desktop positioning)
- `canSendMessage` - Whether to show the "Send Message" button
- `onSendMessage` - Callback when "Send Message" is clicked
- `onClose` - Callback to close the popover/sheet
-->
<script lang="ts">
  import type { PresenceStatus } from '$lib/gql/graphql';
  import UserAvatar from '$lib/components/UserAvatar.svelte';
  import ContextMenu from '$lib/ui/ContextMenu.svelte';
  import { Pill } from '$lib/ui';
  import { getLiveDisplayName, getLiveLogin } from '$lib/state/userProfiles.svelte';

  let {
    user,
    anchorRect,
    canSendMessage = false,
    onSendMessage,
    onClose
  }: {
    user: {
      id: string;
      login: string;
      displayName: string;
      avatarUrl?: string | null;
      presenceStatus: PresenceStatus;
      isBot?: boolean;
    };
    anchorRect?: { top: number; bottom: number; left: number } | null;
    canSendMessage?: boolean;
    onSendMessage?: () => void;
    onClose?: () => void;
  } = $props();

  const displayName = $derived(getLiveDisplayName(user.id, user.displayName || user.login));

  function handleSendMessage() {
    onSendMessage?.();
    onClose?.();
  }
</script>

<ContextMenu
  anchor={anchorRect}
  role="dialog"
  ariaLabel="User profile"
  class="w-64"
  onclose={() => onClose?.()}
>
  <div class="rounded-md bg-background">
    <div class="flex items-center gap-3 p-3">
      <UserAvatar {user} size="md" />
      <div class="min-w-0 flex-1">
        <div class="flex min-w-0 items-center gap-2">
          <div class="truncate font-semibold">{displayName}</div>
          {#if user.isBot}
            <Pill>Bot</Pill>
          {/if}
        </div>
        <div class="truncate text-xs text-muted">@{getLiveLogin(user.id, user.login)}</div>
      </div>
    </div>

    {#if canSendMessage}
      <div class="border-t border-border p-1">
        <button type="button" class="sidebar-item" onclick={handleSendMessage}>
          Send Message
        </button>
      </div>
    {/if}
  </div>
</ContextMenu>
