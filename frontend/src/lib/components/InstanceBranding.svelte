<script lang="ts">
  import { renderMarkdown } from '$lib/markdown';

  let {
    name,
    iconUrl = null,
    ogImageUrl = null,
    welcomeMessage = null,
  }: {
    name: string;
    iconUrl?: string | null;
    ogImageUrl?: string | null;
    welcomeMessage?: string | null;
  } = $props();
</script>

<div class="flex flex-col items-center gap-5">
  <div class="flex items-center gap-4">
    {#if iconUrl}
      <img src={iconUrl} alt="" class="h-12 w-12 rounded-xl" />
    {/if}

    <h3 class="text-2xl font-bold">{name}</h3>
  </div>

  {#if ogImageUrl}
    <img src={ogImageUrl} alt="" class="w-full rounded-lg" />
  {/if}

  {#if welcomeMessage}
    <div class="w-full rounded-lg border border-border bg-surface p-4 prose max-w-none text-muted">
      {#await renderMarkdown(welcomeMessage)}
        <p>{welcomeMessage}</p>
      {:then html}
        <!-- eslint-disable-next-line svelte/no-at-html-tags -- admin-configured content -->
        {@html html}
      {/await}
    </div>
  {/if}
</div>
