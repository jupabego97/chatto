---
name: chatto-admin-frontend
description: "Admin frontend development patterns for instance admin and space settings pages. Covers Panel component, DataTable, UI patterns, implicit roles, and page layout."
---

# Admin Frontend

## Structure

- Admin routes live under `/chat/admin/` (integrated into the chat layout, similar to `/chat/settings/`)
- Admin layout (`routes/chat/admin/+layout.svelte`) handles permission checks and access denied states
- Sidebar navigation for admin sections (with "Back to Chat" link at bottom)
- Link to admin in chat header (only visible to admins)

## Panel Component Scope

The `Panel` component from `$lib/components/admin` is used in **both** instance admin pages AND space settings pages. This ensures visual consistency across all administrative interfaces:

- **Instance admin** (`/chat/admin/*`): System-wide configuration
- **Space settings** (`/chat/[spaceId]/settings/*`): Per-space configuration

When updating the Panel component, remember that changes affect both areas.

## UI Component Patterns

Admin pages follow consistent patterns using shared components from `$lib/components/admin`:

### Data Tables

Wrap `DataTable` with `Panel noPadding` for consistent styling:

```svelte
<Panel noPadding>
  <DataTable items={entries} columns={4} emptyMessage="No data yet">
    {#snippet header()}
      <th class="px-4 py-3 font-medium">Column 1</th>
      <!-- ... -->
    {/snippet}
    {#snippet row(item)}
      <td class="px-4 py-3">...</td>
      <!-- ... -->
    {/snippet}
  </DataTable>
</Panel>
```

### Error States

Use consistent error styling:

```svelte
{#if error}
  <div class="rounded-lg border border-danger/20 bg-danger/10 p-4 text-danger">
    {error}
  </div>
{/if}
```

### Loading States

Simple text indicator:

```svelte
{:else if loading}
  <div class="text-muted">Loading...</div>
{/if}
```

### Item Counts

Show total count below data tables:

```svelte
<div class="text-sm text-muted">{items.length} item(s)</div>
```

### Page Layout

Use `PaneHeader` for title/subtitle, then content in a scrollable container:

```svelte
<PaneHeader title="Page Title" subtitle="Description" />

<div class="flex flex-col gap-6 overflow-y-auto p-6">
  <!-- Content -->
</div>
```

## Implicit vs Explicit Roles

Some roles are **implicit** (automatically assigned based on user state) and should not be editable:

| Role       | Scope    | Condition                  | UI Treatment                       |
| ---------- | -------- | -------------------------- | ---------------------------------- |
| `everyone` | Instance | All authenticated users    | Always checked, disabled           |
| `verified` | Instance | User has verified email(s) | Checked if condition met, disabled |
| `everyone` | Space    | All space members          | Always checked, disabled           |

When building role assignment UI:

```typescript
// Instance admin: everyone/verified are implicit
const isImplicitRole = (roleName: string) =>
  roleName === "everyone" || roleName === "verified";

const hasImplicitRole = (roleName: string) => {
  if (roleName === "everyone") return true;
  if (roleName === "verified") return (user?.verifiedEmails?.length ?? 0) > 0;
  return false;
};

// Space admin: everyone is implicit (all space members have this role)
const IMPLICIT_ROLES = ["everyone"];
const isImplicitRole = (roleName: string) => IMPLICIT_ROLES.includes(roleName);
```

Show implicit roles as checked and disabled with explanatory text like "(automatic)" or "Implicit for all members".
