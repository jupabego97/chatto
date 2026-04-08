<!--
@component

A reusable floating menu/popover. On desktop, positions itself at a viewport point or anchored to
an element. On touch devices, renders as a BottomSheet instead. Handles click-outside dismissal,
Escape key, and scroll dismissal (desktop), or swipe-to-close (mobile).

**Props:**
- `position` - Viewport coordinates {x, y} for point-based positioning (context menus)
- `anchor` - Element rect {top, bottom, left} for anchor-based positioning (popovers)
- `role` - ARIA role (default: "menu")
- `ariaLabel` - ARIA label for the container
- `class` - Additional CSS classes for the outer container (desktop only)
- `onclose` - Callback when the menu should be dismissed

On desktop, exactly one of `position` or `anchor` must be provided. On touch devices, both are
ignored (the BottomSheet handles its own positioning).
-->
<script lang="ts">
  import { fade } from 'svelte/transition';
  import type { Snippet } from 'svelte';
  import BottomSheet from './BottomSheet.svelte';
  import { isTouchDevice } from '$lib/utils/isTouchDevice';

  const PADDING = 8; // Min distance from viewport edge
  const GAP = 4; // Space between anchor and popover (anchor mode only)

  let {
    position,
    anchor,
    role = 'menu',
    ariaLabel,
    class: className,
    onclose,
    onmouseenter,
    onmouseleave,
    children
  }: {
    position?: { x: number; y: number; alignRight?: boolean; centerX?: boolean };
    anchor?: { top: number; bottom: number; left: number } | null;
    role?: string;
    ariaLabel?: string;
    class?: string;
    onclose: () => void;
    onmouseenter?: () => void;
    onmouseleave?: () => void;
    children: Snippet;
  } = $props();

  const isTouch = isTouchDevice();
  let sheetVisible = $state(true);

  /**
   * Attachment that positions the menu and adds dismissal behavior.
   * Supports two modes:
   * - Point mode (position): places at exact cursor coordinates, flips near edges
   * - Anchor mode (anchor): places below/above an element rect, aligns left
   */
  function positionMenu(node: HTMLElement) {
    // Show as a popover so it renders in the browser's top layer, escaping
    // CSS containment boundaries (e.g., virtua's contain:layout on list items
    // which makes position:fixed relative to the container instead of the viewport).
    node.showPopover();

    const { height, width } = node.getBoundingClientRect();
    let top: number;
    let left: number;

    if (anchor) {
      // Anchor mode: prefer below element, fall back above, then pin to bottom
      if (anchor.bottom + GAP + height <= window.innerHeight - PADDING) {
        top = anchor.bottom + GAP;
      } else if (anchor.top - GAP - height >= PADDING) {
        top = anchor.top - GAP - height;
      } else {
        top = Math.max(PADDING, window.innerHeight - PADDING - height);
      }

      // Align with anchor left, clamp to viewport
      left = anchor.left;
      left = Math.max(PADDING, Math.min(left, window.innerWidth - PADDING - width));
    } else if (position) {
      // Point mode: prefer below/right of cursor, flip if near edges
      if (position.y + height <= window.innerHeight - PADDING) {
        top = position.y;
      } else if (position.y - height >= PADDING) {
        top = position.y - height;
      } else {
        top = Math.max(PADDING, window.innerHeight - PADDING - height);
      }

      if (position.centerX) {
        // Center-aligned: menu centered horizontally on x
        left = position.x - width / 2;
        left = Math.max(PADDING, Math.min(left, window.innerWidth - PADDING - width));
      } else if (position.alignRight) {
        // Right-aligned: menu's right edge at x, extending left
        left = position.x - width;
        left = Math.max(PADDING, Math.min(left, window.innerWidth - PADDING - width));
      } else if (position.x + width <= window.innerWidth - PADDING) {
        left = position.x;
      } else {
        left = Math.max(PADDING, position.x - width);
      }
    } else {
      return;
    }

    node.style.top = `${top}px`;
    node.style.left = `${left}px`;

    // Click-outside dismissal (delayed one frame to avoid catching the opening click)
    function handlePointerDown(e: PointerEvent) {
      if (node.contains(e.target as Node)) return;
      onclose();
    }

    // Scroll dismissal (capture phase catches scrolling in any container)
    // Ignore scroll events from inside the menu (e.g., emoji picker grid)
    function handleScroll(e: Event) {
      if (node.contains(e.target as Node)) return;
      onclose();
    }

    const frame = requestAnimationFrame(() => {
      document.addEventListener('pointerdown', handlePointerDown);
      window.addEventListener('scroll', handleScroll, { capture: true });
    });

    return () => {
      cancelAnimationFrame(frame);
      document.removeEventListener('pointerdown', handlePointerDown);
      window.removeEventListener('scroll', handleScroll, { capture: true });
    };
  }

  function handleKeydown(e: KeyboardEvent) {
    if (isTouch) return;
    if (e.key === 'Escape') {
      e.stopPropagation();
      onclose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isTouch}
  <BottomSheet bind:visible={sheetVisible} {onclose}>
    {@render children()}
  </BottomSheet>
{:else}
  <div
    {@attach positionMenu}
    popover="manual"
    transition:fade|global={{ duration: 100 }}
    class={['menu fixed z-50 min-w-48', className]}
    {role}
    aria-label={ariaLabel}
    {onmouseenter}
    {onmouseleave}
  >
    <div class="flex flex-col gap-1">
      {@render children()}
    </div>
  </div>
{/if}
