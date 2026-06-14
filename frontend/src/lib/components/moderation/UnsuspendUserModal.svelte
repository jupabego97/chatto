<script lang="ts">
  import type { PresenceStatus } from '$lib/gql/graphql';
  import UserAvatar from '$lib/components/UserAvatar.svelte';
  import FormDialog from '$lib/ui/FormDialog.svelte';
  import { TextArea } from '$lib/ui/form';

  type User = {
    id: string;
    login: string;
    displayName: string;
    avatarUrl?: string | null;
    presenceStatus: PresenceStatus;
  };

  let {
    user = null,
    userId,
    submitting = false,
    error = null,
    onconfirm,
    onclose
  }: {
    user?: User | null;
    userId: string;
    submitting?: boolean;
    error?: string | null;
    onconfirm?: (reason: string) => void;
    onclose?: () => void;
  } = $props();

  let visible = $state(true);
  let reason = $state('');

  const displayName = $derived(user?.displayName || user?.login || userId);
  const disabled = $derived(reason.trim().length === 0 || submitting);

  function handleSubmit() {
    if (disabled) return;
    onconfirm?.(reason.trim());
  }
</script>

<FormDialog
  bind:visible
  title={`Unsuspend ${displayName}`}
  size="sm"
  submitLabel="Unsuspend"
  submitTone="warning"
  submitIcon="iconify uil--user-check"
  submitLoadingText="Unsuspending..."
  loading={submitting}
  {disabled}
  {error}
  onsubmit={handleSubmit}
  onclose={() => onclose?.()}
>
  <div class="flex items-center gap-3 rounded-md border border-border bg-surface-100 p-3">
    {#if user}
      <UserAvatar {user} size="md" />
    {:else}
      <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-surface-200 text-muted">
        <span class="iconify text-lg uil--user"></span>
      </div>
    {/if}
    <div class="min-w-0 flex-1">
      <div class="truncate font-medium text-text">{displayName}</div>
      {#if user}
        <div class="truncate text-sm text-muted">@{user.login}</div>
      {:else}
        <div class="truncate text-sm text-muted">{userId}</div>
      {/if}
    </div>
  </div>

  <TextArea
    id="unsuspend-user-reason"
    label="Reason"
    bind:value={reason}
    rows={4}
    maxlength={1000}
    required
    disabled={submitting}
  />
</FormDialog>
