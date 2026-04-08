<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    children,
    visible = $bindable(false),
    onclose
  }: {
    visible?: boolean;
    children: Snippet;
    onclose?: () => void;
  } = $props();

  let dialogEl: HTMLDialogElement | undefined = $state();
  let closing = $state(false);

  $effect(() => {
    if (visible) {
      closing = false;
      dialogEl?.showModal();
    } else if (dialogEl?.open && !closing) {
      dialogEl?.close();
    }
  });

  function handleNativeClose() {
    visible = false;
    closing = false;
    onclose?.();
  }

  function close() {
    if (!dialogEl?.open || closing) return;
    closing = true;
    // Wait for exit animation, then close
    setTimeout(() => {
      dialogEl?.close();
    }, 200);
  }
</script>

<dialog
  bind:this={dialogEl}
  onclose={handleNativeClose}
  oncancel={(e) => {
    e.preventDefault();
    // On Android, the virtual keyboard appearance can fire a spurious cancel event.
    // Don't close if an input/textarea inside the dialog currently has focus.
    const active = document.activeElement;
    if (active && dialogEl?.contains(active) && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA')) {
      return;
    }
    close();
  }}
  onclick={(e) => {
    // Use coordinate check instead of e.target === dialogEl.
    // On mobile, tapping an input (eg. emoji search) triggers the virtual keyboard,
    // which shifts content during touch-to-click synthesis. This can cause e.target
    // to resolve to the dialog element instead of the input, dismissing the sheet.
    const content = dialogEl?.firstElementChild as HTMLElement | null;
    if (!content) return;
    const rect = content.getBoundingClientRect();
    if (e.clientY < rect.top) close();
  }}
  class="bottom-sheet m-0 mt-auto w-full max-w-full bg-transparent p-0 backdrop:bg-black/50"
  class:closing
>
  <div class="pb-safe rounded-t-xl border-t border-border bg-surface">
    <!-- Drag handle - click to close -->
    <button
      type="button"
      class="flex w-full cursor-pointer justify-center py-3"
      onclick={close}
      aria-label="Close"
    >
      <div class="h-1 w-10 rounded-full bg-muted/40"></div>
    </button>

    <!-- Content -->
    <div class="px-4 pb-4">
      {@render children()}
    </div>
  </div>
</dialog>

<style>
  dialog.bottom-sheet[open] {
    animation: slide-up 200ms ease-out;
  }

  dialog.bottom-sheet[open]::backdrop {
    animation: backdrop-fade-in 200ms ease-out;
  }

  dialog.bottom-sheet[open].closing {
    animation: slide-down 200ms ease-in forwards;
  }

  dialog.bottom-sheet[open].closing::backdrop {
    animation: backdrop-fade-out 200ms ease-in forwards;
  }

  @keyframes slide-up {
    from {
      transform: translateY(100%);
    }
    to {
      transform: translateY(0);
    }
  }

  @keyframes slide-down {
    from {
      transform: translateY(0);
    }
    to {
      transform: translateY(100%);
    }
  }

  @keyframes backdrop-fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes backdrop-fade-out {
    from {
      opacity: 1;
    }
    to {
      opacity: 0;
    }
  }
</style>
