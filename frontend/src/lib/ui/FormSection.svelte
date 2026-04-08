<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    title,
    maxWidth,
    bordered = false,
    children
  }: {
    title: string;
    maxWidth?: 'max-w-md' | 'max-w-lg' | 'max-w-xl' | 'max-w-2xl';
    bordered?: boolean;
    children: Snippet;
  } = $props();
</script>

<section class={[maxWidth, bordered && 'border-t border-border pt-6']}>
  <h3 class="mb-4 text-sm font-semibold text-muted uppercase">{title}</h3>
  {@render children()}
</section>
