<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    label,
    id,
    error,
    description,
    required = false,
    children
  }: {
    label: string;
    id?: string;
    error?: string;
    description?: string;
    required?: boolean;
    children: Snippet;
  } = $props();
</script>

<div>
  <label for={id} class="mb-1 block text-sm font-medium text-muted">
    {label}
    {#if required}<span class="text-error">*</span>{/if}
  </label>

  {@render children()}

  {#if error}
    <p id="{id}-error" class="mt-1 text-xs text-error">{error}</p>
  {:else if description}
    <p id="{id}-description" class="mt-1 text-xs text-muted">{description}</p>
  {/if}
</div>
