---
# chatto-5ko7
title: Add sidebar-item-active utility class
status: todo
type: task
priority: normal
created_at: 2026-01-22T17:27:27Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzs
parent: chatto-2qkh
---

Add a utility class for the active sidebar item state to centralize the styling.

## Current State
Active state styling is repeated across multiple components:
```svelte
class={['sidebar-item', isActive(item.href) ? 'bg-surface-100' : '']}
```

## Proposed Change
Add to `frontend/src/app.css`:

```css
@utility sidebar-item-active {
  @apply bg-surface-100;
}
```

Then usage becomes:
```svelte
class={['sidebar-item', isActive(item.href) && 'sidebar-item-active']}
```

## Benefits
- Single place to update active styling
- More semantic class name
- Consistent with existing utility pattern (`sidebar-item`, `sidebar-icon`)

## Tasks
- [ ] Add `sidebar-item-active` utility to app.css
- [ ] Update usages across the codebase (or defer to SidebarNav component work)
