---
# chatto-ta3q
title: Consolidate SpaceList icon button styling
status: todo
type: task
priority: normal
created_at: 2026-01-22T17:27:48Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzV
parent: chatto-2qkh
---

The DM, Explore Spaces, and Admin buttons in SpaceList share nearly identical styling that could be consolidated.

## Current State
In `frontend/src/lib/SpaceList.svelte`, these buttons share the same pattern:

```svelte
<a
  href="..."
  title="..."
  class={[
    'flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-3xl transition-colors duration-100',
    isActive
      ? 'bg-surface-highlighted text-text'
      : 'bg-surface text-text/70 hover:bg-surface-highlighted hover:text-text'
  ]}
>
  <span class="iconify ..."></span>
</a>
```

## Options

### Option A: Extract shared class string
Create a helper function or constant for the base classes:
```typescript
const iconButtonClasses = (active: boolean) => [
  'flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-3xl transition-colors duration-100',
  active ? 'bg-surface-highlighted text-text' : 'bg-surface text-text/70 hover:bg-surface-highlighted hover:text-text'
];
```

### Option B: Create SpaceListIconLink component
Similar to SpaceIcon but for link-style buttons:
```svelte
<SpaceListIconLink href="/chat/dm" title="Direct Messages" icon="uil--comment-alt-lines" active={isDMActive} />
```

### Option C: Add utility classes
Add to app.css:
```css
@utility space-list-icon {
  @apply flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-3xl transition-colors duration-100;
  @apply bg-surface text-text/70 hover:bg-surface-highlighted hover:text-text;
}
@utility space-list-icon-active {
  @apply bg-surface-highlighted text-text;
}
```

## Recommendation
Option B (component) provides the best balance of reusability and type safety, but Option A (helper function) is simpler and doesn't require a new file.

## Tasks
- [ ] Decide on approach
- [ ] Implement chosen solution
- [ ] Update SpaceList.svelte
