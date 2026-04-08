<script lang="ts">
  import { getToasts, toast } from './toastState.svelte';
  import Toast from './Toast.svelte';

  const toasts = $derived(getToasts());
</script>

{#if toasts.length > 0}
  <div class="fixed right-4 bottom-4 z-50 flex flex-col gap-2">
    {#each toasts as t (t.id)}
      <div class="toast-enter">
        <Toast
          type={t.type}
          message={t.message}
          action={t.action}
          onDismiss={() => toast.remove(t.id)}
        />
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-enter {
    animation: slide-in 150ms ease-out;
  }

  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateX(100%);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }
</style>
