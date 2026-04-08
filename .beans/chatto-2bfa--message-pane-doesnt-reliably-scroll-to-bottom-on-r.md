---
# chatto-2bfa
title: Message pane doesn't reliably scroll to bottom on room load
status: todo
type: bug
priority: normal
created_at: 2026-01-18T23:12:07Z
updated_at: 2026-03-18T05:34:10Z
order: zzw
---

## Bug Description

When loading Chatto and navigating to a room, the message pane doesn't always scroll all the way to the bottom. Users expect to see the most recent messages immediately.

## Expected Behavior

When entering a room, the message list should scroll to the bottom showing the latest messages.

## Actual Behavior

Sometimes the scroll position is partway up the message list, requiring manual scrolling to see recent messages.

## Likely Causes

- Race condition between message rendering and scroll-to-bottom call
- Images/embeds loading after initial scroll, changing content height
- Async Svelte rendering completing after scroll attempt
- Virtual scrolling (if used) not fully initialized before scroll

## Investigation Steps

- [ ] Identify where scroll-to-bottom is triggered on room load
- [ ] Check timing relative to message list rendering
- [ ] Test with/without images and embeds
- [ ] Check if issue is worse on slow connections or large rooms

## Potential Fixes

- Use `tick()` to wait for DOM update before scrolling
- Use ResizeObserver or MutationObserver to re-scroll after content changes
- Scroll after images load (`onload` handlers)
- Multiple scroll attempts with slight delays as fallback
- Ensure scroll happens after any async data fetching completes
