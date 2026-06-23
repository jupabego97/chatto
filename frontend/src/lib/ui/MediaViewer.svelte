<script lang="ts">
  import { onMount } from 'svelte';
  import { useNativeZoomAllowance } from '$lib/hooks';
  import type { MediaViewerItem } from './mediaViewer';

  import 'vidstack/player/styles/default/theme.css';
  import 'vidstack/player/styles/default/layouts/video.css';

  let {
    items,
    index = $bindable(0),
    onclose
  }: {
    items: MediaViewerItem[];
    index?: number;
    onclose: () => void;
  } = $props();

  let elementsReady = $state(false);
  let dialogEl: HTMLDialogElement | undefined = $state();
  let nativeVideoEl: HTMLVideoElement | null = $state(null);
  let playerEl: HTMLElement | null = $state(null);

  const current = $derived(items[index]);
  const hasMultiple = $derived(items.length > 1);
  const nativeZoomEnabled = $derived(current?.kind === 'image');
  const openHref = $derived(current ? (current.openUrl ?? current.src) : null);

  useNativeZoomAllowance(() => nativeZoomEnabled);

  onMount(async () => {
    await Promise.all([
      import('vidstack/player'),
      import('vidstack/player/layouts'),
      import('vidstack/player/ui')
    ]);
    elementsReady = true;
  });

  $effect(() => {
    dialogEl?.showModal();
  });

  $effect(() => {
    if (current?.kind !== 'video' || !current.startTime) return;

    const startTime = current.startTime;
    const video = current.autoLoop ? nativeVideoEl : playerEl?.querySelector('video');
    if (!video) return;
    const videoEl = video;

    function seek() {
      videoEl.currentTime = startTime;
    }

    if (videoEl.readyState >= 1) {
      seek();
      return;
    }

    videoEl.addEventListener('loadedmetadata', seek, { once: true });
    return () => videoEl.removeEventListener('loadedmetadata', seek);
  });

  function close() {
    onclose();
  }

  function navigate(direction: -1 | 1) {
    index = (index + direction + items.length) % items.length;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      close();
    } else if (e.key === 'ArrowLeft' && hasMultiple) {
      e.preventDefault();
      navigate(-1);
    } else if (e.key === 'ArrowRight' && hasMultiple) {
      e.preventDefault();
      navigate(1);
    }
  }

  function openCurrentSource() {
    if (!openHref) return;
    window.open(openHref, '_blank', 'noopener,noreferrer');
  }
</script>

<dialog
  bind:this={dialogEl}
  onclose={close}
  onkeydown={handleKeydown}
  onclick={(e) => {
    if (e.target === dialogEl) close();
  }}
  class="fixed inset-0 m-0 flex h-dvh max-h-dvh w-dvw max-w-dvw items-center justify-center border-none bg-black/90 p-0 text-white backdrop:bg-transparent"
>
  <button
    type="button"
    class="absolute inset-0 cursor-default"
    onclick={close}
    aria-label="Close media viewer backdrop"
  ></button>

  {#if current}
    <div class="pointer-events-none relative flex h-full w-full flex-col">
      <div class="flex min-h-0 flex-1 items-center justify-center gap-2 p-3 sm:p-5">
        {#if hasMultiple}
          <button
            type="button"
            onclick={() => navigate(-1)}
            class="pointer-events-auto flex h-10 w-10 shrink-0 cursor-pointer items-center justify-center rounded-full text-white/70 transition-colors hover:bg-white/10 hover:text-white focus-visible:bg-white/10 focus-visible:text-white"
            aria-label="Previous media"
          >
            <span class="iconify text-3xl uil--angle-left-b"></span>
          </button>
        {/if}

        <div class="flex min-h-0 min-w-0 flex-1 items-center justify-center">
          {#if current.kind === 'image'}
            <img
              src={current.src}
              alt={current.alt ?? current.filename ?? 'Image'}
              class="pointer-events-auto max-h-[calc(100dvh-5.5rem)] max-w-full object-contain"
            />
          {:else if current.autoLoop}
            <video
              bind:this={nativeVideoEl}
              autoplay
              loop
              muted
              playsinline
              data-autoloop
              src={current.src}
              poster={current.poster ?? undefined}
              class="pointer-events-auto max-h-[calc(100dvh-5.5rem)] max-w-full object-contain"
              style:aspect-ratio={current.width && current.height ? `${current.width} / ${current.height}` : undefined}
            >
              <track kind="captions" />
            </video>
          {:else if elementsReady}
            <media-player
              bind:this={playerEl}
              src={{ src: current.src, type: 'video/mp4' }}
              playsinline
              class="pointer-events-auto max-h-[calc(100dvh-5.5rem)] w-full max-w-[min(100%,90rem)]"
              style:aspect-ratio={current.width && current.height ? `${current.width} / ${current.height}` : '16 / 9'}
            >
              <media-provider>
                {#if current.poster}
                  <media-poster class="vds-poster" src={current.poster} alt={current.filename ?? 'Video'}
                  ></media-poster>
                {/if}
              </media-provider>
              <media-video-layout></media-video-layout>
            </media-player>
          {/if}
        </div>

        {#if hasMultiple}
          <button
            type="button"
            onclick={() => navigate(1)}
            class="pointer-events-auto flex h-10 w-10 shrink-0 cursor-pointer items-center justify-center rounded-full text-white/70 transition-colors hover:bg-white/10 hover:text-white focus-visible:bg-white/10 focus-visible:text-white"
            aria-label="Next media"
          >
            <span class="iconify text-3xl uil--angle-right-b"></span>
          </button>
        {/if}
      </div>

      <div class="pointer-events-auto flex min-h-16 flex-wrap items-center justify-center gap-x-4 gap-y-2 px-4 pb-4 text-white/80">
        {#if current.filename}
          <span class="max-w-[min(70vw,36rem)] truncate text-sm">{current.filename}</span>
        {/if}

        {#if hasMultiple}
          <span class="text-sm text-white/50">{index + 1} / {items.length}</span>
        {/if}

        {#if openHref}
          <button
            type="button"
            onclick={openCurrentSource}
            class="flex items-center gap-1 text-sm text-white/60 hover:text-white"
          >
            <span class="iconify uil--external-link-alt"></span>
            {current.kind === 'image' ? 'Open original' : 'Open source'}
          </button>
        {/if}
      </div>

      <button
        type="button"
        onclick={close}
        class="pointer-events-auto absolute top-3 right-3 flex h-10 w-10 cursor-pointer items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
        aria-label="Close media viewer"
      >
        <span class="iconify text-2xl uil--times"></span>
      </button>
    </div>
  {/if}
</dialog>

<style>
  dialog[open] {
    animation: fade-in 150ms ease-out;
  }

  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  :global(media-player .vds-settings-menu),
  :global(media-player .vds-chapters-menu),
  :global(media-player .vds-fullscreen-button) {
    display: none !important;
  }
</style>
