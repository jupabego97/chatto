---
# chatto-200l
title: Create SettingsLayout shell component
status: todo
type: task
priority: normal
created_at: 2026-01-22T17:27:15Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzs
parent: chatto-2qkh
---

Extract the shared settings layout structure into a reusable component.

## Current State
Three layouts share nearly identical structure:
- `frontend/src/routes/chat/admin/+layout.svelte`
- `frontend/src/routes/chat/settings/+layout.svelte`
- `frontend/src/routes/chat/[spaceId]/+layout.svelte` (settings mode)

Each has:
- Frame wrapper with connection state
- Sidebar with PaneHeader
- Navigation items
- "Back to X" link at bottom
- Content slot

## Proposed Component
Create `$lib/components/SettingsLayout.svelte`:

```svelte
<script lang="ts">
  import Frame from '$lib/Frame.svelte';
  import PaneHeader from '$lib/PaneHeader.svelte';
  import SidebarNav from './SidebarNav.svelte';
  
  type NavItem = { href: string; label: string; icon: string };
  
  let {
    title,
    subtitle,
    navItems,
    backHref,
    backLabel,
    children
  }: {
    title: string;
    subtitle?: string;
    navItems: NavItem[];
    backHref: string;
    backLabel: string;
    children: Snippet;
  } = $props();
</script>

<Frame>
  <aside class="w-56 flex-col border-r border-border md:flex">
    <PaneHeader {title} {subtitle} />
    <SidebarNav items={navItems} />
    <div class="border-t border-border p-2">
      <a href={backHref} class="sidebar-item text-muted">
        <span class="sidebar-icon iconify uil--arrow-left"></span>
        {backLabel}
      </a>
    </div>
  </aside>
  {@render children()}
</Frame>
```

## Tasks
- [ ] Create `SettingsLayout.svelte` component
- [ ] Depends on: SidebarNav component (chatto-gxws)
- [ ] Update admin layout to use component
- [ ] Update settings layout to use component
- [ ] Update space settings layout to use component
- [ ] Handle mobile nav toggle (currently varies by layout)
