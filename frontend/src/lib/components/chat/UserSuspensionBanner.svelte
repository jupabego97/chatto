<script lang="ts">
  import { getServerPermissions } from '$lib/state/server/permissions.svelte';
  import { getUserSettings } from '$lib/state/userSettings.svelte';
  import { formatDate as formatDateUtil } from '$lib/utils/formatTime';

  const serverPerms = getServerPermissions();
  const userSettings = getUserSettings();

  const suspension = $derived(serverPerms.current.suspension);
  const isSuspended = $derived(suspension.isSuspended);
  const expiresAt = $derived(suspension.expiresAt ?? null);
  const expiryLabel = $derived(
    expiresAt ? `until ${formatDateUtil(expiresAt, userSettings)}` : 'indefinitely'
  );
</script>

{#if isSuspended}
  <div class="border-b border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger">
    <div class="mx-auto flex max-w-screen-2xl items-center gap-3">
      <span class="iconify uil--exclamation-triangle shrink-0 text-lg"></span>
      <div class="min-w-0">
        <div class="font-medium">Your account is suspended {expiryLabel}.</div>
        <div class="text-danger/80">
          Posting, reactions, room access changes, and administration are temporarily unavailable.
        </div>
      </div>
    </div>
  </div>
{/if}
