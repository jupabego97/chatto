<script lang="ts">
  import type { Snippet } from 'svelte';
  import Dialog from './Dialog.svelte';

  let {
    children,
    visible = $bindable(true),
    title,
    actionLabel = 'Confirm',
    actionIcon = 'iconify uil--check',
    loading = false,
    onconfirm,
    onclose
  }: {
    children: Snippet;
    visible?: boolean;
    title: string;
    actionLabel?: string;
    actionIcon?: string;
    loading?: boolean;
    onconfirm: () => void;
    onclose: () => void;
  } = $props();
</script>

<Dialog {visible} {title} size="sm" {onclose}>
  <p class="mb-4 text-muted">
    {@render children()}
  </p>
  <div class="flex justify-end gap-3">
    <button
      type="button"
      class="flex cursor-pointer items-center gap-2 rounded-lg bg-surface-200 px-4 py-2 text-sm font-medium text-text hover:bg-surface-300"
      onclick={onclose}
      disabled={loading}
    >
      <span class="iconify uil--times"></span>
      Cancel
    </button>
    <button
      type="button"
      class="flex cursor-pointer items-center gap-2 rounded-lg bg-danger px-4 py-2 text-sm font-medium text-white hover:bg-danger/90 disabled:opacity-50"
      onclick={onconfirm}
      disabled={loading}
    >
      <span class={actionIcon}></span>
      {loading ? `${actionLabel}...` : actionLabel}
    </button>
  </div>
</Dialog>
