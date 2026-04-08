---
# chatto-9xwm
title: Draggable pane dividers for room/thread layout
status: todo
type: feature
priority: normal
created_at: 2026-01-18T22:48:23Z
updated_at: 2026-03-18T05:34:10Z
order: zzzz
---

## Overview

Make the divider between the room pane and thread pane draggable, allowing users to resize panes to their preference.

## Scope

- Divider between main room content and thread side pane
- Potentially also sidebar divider (if not already resizable)

## Implementation

- Add drag handle/divider component
- Track mouse/touch drag events to resize panes
- Set min/max width constraints to prevent unusable sizes
- Persist user's preferred width to localStorage
- Restore saved width on page load

## UX Details

- Visual affordance on hover (cursor change, highlight)
- Smooth resize during drag
- Double-click to reset to default width
- Respect minimum widths for readability
- Handle edge case: thread pane closed vs minimized vs resized to min

## Accessibility

- Keyboard support for resizing (arrow keys when focused)
- Appropriate ARIA attributes for the separator
