<script lang="ts">
  import { Button } from '$lib/ui/form';
  import Dialog from '$lib/ui/Dialog.svelte';

  let {
    roleDisplayName,
    deleting = false,
    onConfirm,
    onCancel
  }: {
    roleDisplayName: string;
    deleting?: boolean;
    onConfirm: () => void;
    onCancel: () => void;
  } = $props();

  let visible = $state(true);

  function handleClose() {
    visible = false;
    onCancel();
  }
</script>

<Dialog {visible} title="Delete Role" size="sm" onclose={handleClose}>
  <p class="mb-4 text-muted">
    Are you sure you want to delete the role <strong>{roleDisplayName}</strong>? This will:
  </p>
  <ul class="mb-4 list-inside list-disc text-sm text-muted">
    <li>Remove the role from all users who have it</li>
    <li>Delete all permission grants for this role</li>
  </ul>
  <p class="text-sm font-medium text-error">This action cannot be undone.</p>

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button variant="secondary" onclick={handleClose} disabled={deleting}>Cancel</Button>
      <Button variant="danger" onclick={onConfirm} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete Role'}
      </Button>
    </div>
  {/snippet}
</Dialog>
