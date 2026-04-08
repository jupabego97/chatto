---
name: chatto-frontend
description: "Frontend development guidance for Svelte 5, SvelteKit, and the Chatto UI. Covers runes/reactivity, state management, event buses, GraphQL client, components, layouts, modals, form validation, CSS patterns, and PWA setup."
---

# Frontend

## Browser Support

Target modern evergreen browsers only. No support for Internet Explorer or legacy browsers.

- Chrome/Edge 90+ (2021)
- Firefox 90+ (2021)
- Safari 14+ (2020)

Use modern Web APIs (`navigator.clipboard`, `fetch`, `async/await`, etc.) without fallbacks.

## Formatting

- Svelte files in this project use tabs for indentation. Always match the existing indentation style of the file being edited. Before editing, check whether the file uses tabs or spaces.

## Tooling

- Use `pnpm` for package management and running scripts

## Type Generation

Frontend TypeScript types for permission constants and other shared types are generated from Go using `tygo`. When Go types are renamed or added:

1. Run `mise codegen-types` to regenerate `frontend/src/lib/types/core.ts`
2. Update all frontend imports that reference the changed types
3. The TypeScript compiler will catch mismatches, but e2e tests may fail with confusing errors (like "Access Denied") if types are stale

**Common symptom**: E2e tests fail with permission errors even though backend tests pass. This often means the frontend is checking for old permission constant values.

## GraphQL Client

- Use `graphqlClient.client` from `$lib/state/graphqlClient.svelte.ts` for mutations/queries
- Do NOT use `getContextClient()` from `@urql/svelte` - the app uses a singleton, not context
- **No GraphQL cache**: The client has no cache exchange - queries always fetch fresh data
- **Event-driven updates**: Instead of relying on cache invalidation, components subscribe to real-time events (via event buses) and update their local state. This fits naturally with the WebSocket-based architecture.
- **Type imports**: Generated GraphQL types are in `$lib/gql/graphql`, not the `$lib/gql` index:
  ```typescript
  import { graphql } from "$lib/gql"; // For the graphql tag function
  import type { SomeType } from "$lib/gql/graphql"; // For generated types
  ```

## Event Buses

- **Instance events** (`instanceEventBus.svelte.ts`): Space/user-level changes - use `onInstanceEvent()`
- **JetStream events** (`spaceEventBus.svelte.ts`): Messages, room events - use `onSpaceEvent()`
- **Live events** (`spaceLiveEventBus.svelte.ts`): Reactions, typing - use `onSpaceLiveEvent()`

**Adding a new instance event handler** (e.g., for `UserProfileUpdatedEvent`):

1. Add the event fields to the subscription in `instanceEventBus.svelte.ts`
2. Create a typed handler hook (e.g., `onUserProfileUpdate`) that filters for that event type
3. Subscribe in components using `$effect(() => onUserProfileUpdate(...))` - the hook returns cleanup

```typescript
// In instanceEventBus.svelte.ts
export function onUserProfileUpdate(handler: UserProfileHandler): () => void {
  const bus = getContext<InstanceEventBus | undefined>(INSTANCE_EVENT_BUS_KEY);
  if (!bus) return () => {}; // Graceful fallback if bus not initialized

  const wrapper: EventHandler = (event) => {
    if (!event.event) return; // Skip unknown event types
    if (event.event.__typename === "UserProfileUpdatedEvent") {
      handler({ userId: event.event.userId, avatarUrl: event.event.avatarUrl });
    }
  };
  bus.handlers.add(wrapper);
  return () => bus.handlers.delete(wrapper);
}

// In component
$effect(() =>
  onUserProfileUpdate((update) => {
    if (update.userId === user.id) liveAvatarUrl = update.avatarUrl;
  }),
);
```

**Whitelist cacheable events in SpaceEventProvider**: When routing subscription events to the message cache, use an explicit whitelist of displayable event types -- NOT a blacklist.

**Event handler null checks**: Always check for `event.event` being null before accessing properties.

## Eagerly Populate Stores Used by Live Event Handlers

Stores that are checked by live event handlers **must be populated at startup**, not lazily on user navigation.

## Real-time Update Patterns

- **Refetch-on-event**: When receiving live events, refetch from server for consistency
- **Inline cache updates**: When the incoming event contains all data needed, update the cache directly

## Centralized Stores

Prefer centralized Svelte stores over duplicating data fetching in multiple components:

- When multiple components need the same data, create a shared store in `$lib/state/`
- Use Svelte 5's `createContext` pattern for scoped stores
- Have a single "owner" component that populates/manages the store lifecycle

**Race condition prevention**: When async operations in `$effect` might complete after dependencies change, capture values at effect start and verify they're still current.

## Modals

- Modals use SvelteKit shallow routing with `pushState('', { modal: { type: '...', ...context } })`
- `App.PageState` in `app.d.ts` defines the type-safe modal state schema
- Close modals with `history.back()`
- Use modals for quick actions (create space, create room)
- Use full pages for settings and discovery/browsing

## Context Menus & Popovers

`ContextMenu` (`$lib/ui/ContextMenu.svelte`) is the unified component for all floating menus and popovers. Automatically renders as floating menu on desktop and BottomSheet on mobile.

Two positioning modes:
- **Point** (`position: { x, y }`) -- for right-click context menus
- **Anchor** (`anchor: { top, bottom, left }`) -- for popovers attached to a trigger element

**Interactive elements inside context menu triggers** need `oncontextmenu` with `preventDefault` + `stopPropagation` and `ontouchstart` with `stopPropagation`.

### CSS Containment and Overlays Inside Virtualized Lists

`virtua` applies `contain: layout` to list items. Use the **Popover API** (`popover="manual"` + `node.showPopover()`) to render overlays in the browser's top layer.

## Form Components

- Use components from `$lib/ui/form` (TextInput, TextArea, Select, Checkbox, Button, FormError)
- Pass `error` prop to inputs for field-level validation display

## Form Validation

- Use Zod for validation - import `z` and `validate` from `$lib/ui/form`
- Define schemas at component level, use `$derived` for reactive error display

```typescript
import { z, validate } from "$lib/ui/form";

const emailSchema = z.email({ error: "Please enter a valid email" });
const emailError = $derived(email ? validate(emailSchema, email) : undefined);
```

## Page Layout

Pages rendered in the main content area must wrap all content in a single flex column container:

```svelte
<div class="flex min-h-0 min-w-0 flex-1 flex-col">
  <PaneHeader title="Page Title" subtitle="Description" />
  <div class="flex flex-col gap-6 overflow-y-auto p-6">
    <!-- Page content -->
  </div>
</div>
```

## UI Guidelines

- Prefer plain HTML + Tailwind utilities over wrapper components
- Use native elements (e.g., `<dialog>` for modals)
- Always add `cursor-pointer` to clickable elements
- **Mobile tap targets**: Use responsive sizing for touch devices

## Permission-Based UI Gating

Query `viewerCan*` fields on the Space type and conditionally render or disable UI.

### Optimistic Permission Checks

For checks that depend on locally available data (like role hierarchy), compute them in the frontend rather than adding backend fields.

## Shared Component Libraries

- **`$lib/components/menus/`** - Context menu content components
- **`$lib/components/rbac/`** - Role management components
- **`$lib/components/admin/`** - Admin panel components
- **`$lib/ui/form/`** - Form components
- **`$lib/ui/`** - Layout components

## CSS Utilities

- We create a library of utility classes in `app.css`
- Use the `skeleton` utility class on `<img>` elements for loading state

## CSS Truncation in Flex Layouts

For `truncate` to work in flexbox, **every flex ancestor** needs `min-w-0`.

## Testing

- When UI elements change, update e2e test selectors accordingly
- Check `frontend/e2e/fixtures/` when modifying shared UI patterns

## PWA Setup

- Manifest at `static/manifest.webmanifest`, linked from `app.html`
- Icons in `static/icons/`
- Use `sharp` dev dependency + `scripts/generate-icons.mjs` to regenerate icons from SVG

---

# Svelte 5

- Always use Svelte 5 idioms, not Svelte 4 patterns (e.g., `<svelte:document>` over legacy event handlers, `$derived` over reactive statements).
- Use runes (`$state`, `$derived`, `$effect`) - no legacy reactivity
- Document components with `@component` in an HTML comment before the `<script>` tag.

  ```svelte
  <!--
  @component

  Brief description of what the component does.

  **Props:**
  - `name` - Description of prop
  -->
  <script lang="ts">
    ...
  </script>
  ```

- Experimental async Svelte enabled - use where appropriate ([docs](https://svelte.dev/docs/svelte/await-expressions))
- Don't put too much code into a single component -- break into smaller components or modules as needed.
- Don't place reusable components or supporting modules in route directories. Shared components belong in `$lib/components/`, and shared state modules in `$lib/state/`.
- Prefer using Tailwind classes over <style> blocks for styling components.
- Keep Tailwind classes in the `class` attribute directly using array syntax: `class={['base', condition ? 'a' : 'b']}`.
- Use Svelte 5's `resolve` from `$app/paths` for typechecked URL paths.
- **`resolve()` exceptions**: Some URLs legitimately cannot use `resolve()`. Use `eslint-disable` comments with clear reasons.
- Use `createContext` for type-safe contexts instead of `setContext`/`getContext` with string keys.

## Svelte 5 Reactivity Pitfalls

### Effect Dependencies Without Unused Variables

Use `void` to avoid ESLint unused variable warnings:

```typescript
$effect(() => {
  void roomId; // Creates dependency, no lint warning
  shouldScrollToBottom = true;
});
```

### Writable `$derived` (Svelte 5.25+)

`$derived` values can be directly overwritten. Perfect for "prop with real-time override" patterns.

**Caution**: Writable `$derived` can cause subtle reversion bugs if the source value changes unexpectedly.

### Tracking Previous Values

**Don't use `$derived` to track previous state** - use `$state` to persist across effect runs.

### Reactive Cache Mutations and Derived Chains

Before mutating a reactive cache, verify the item will survive all downstream filters. If it won't, don't add it.

### Intl.DateTimeFormat Is Expensive

Cache `Intl.DateTimeFormat` instances when calling repeatedly with the same options.

### Method Calls in Templates

Calling store methods directly in templates may not establish reactive dependencies. Wrap in `$derived`.

### {#each} and External State

`{#each}` only re-renders when the iterated array reference changes. For external state dependencies, use `{#key}`.

### Reset State on Entry, Not Cleanup

Reset state in an effect that depends on the new prop value, not in a cleanup function.

### Context and Async Callbacks

`getContext()` can only be called during component initialization. Return update functions with closures during initialization.

### Transitions in Static Conditional Blocks

Use `|global` when a transitioned element is inside a block whose condition is effectively constant.

## State Management Classes

Prefer classes with `$state` fields for stores and state containers. When to use factory functions: Context-scoped stores still need a factory to call `setContext()`, but the factory should instantiate a class.

### SvelteMap/SvelteSet Usage

`SvelteMap` and `SvelteSet` from `svelte/reactivity` are reactive for their method calls. They do NOT need `$state` wrapping. Always use methods instead of reassignment.

## Attachments (`{@attach}`) vs Actions (`use:`)

Svelte 5 replaces `use:action` with `{@attach}` for attaching behavior to DOM elements.

## $effect vs $derived

**`$effect` is an escape hatch** - use `$derived` for computed values, `$effect` only for actual side effects. If you must use `$effect`, extract it into a hook in `$lib/hooks/`.

## SvelteKit Layouts

- **`+layout.svelte` for shared UI**
- **`@render children?.()` for slot content**
- **Active link detection**: Use `page.url.pathname` from `$app/state`

## SvelteKit

- Static SPA only - no server-side rendering (`ssr = false` at root layout)
- Client-side load functions (`+page.ts`, `+layout.ts`) are fine and encouraged
- Never use server load functions (`+page.server.ts`, `+layout.server.ts`)
- Add dependencies as devDependencies (bundled at build time)

### `$app/state` vs `$app/stores`

Prefer `page` from `$app/state` over `$page` from `$app/stores`.

### Refreshing Auth State After Login

After login/registration, use `invalidateAll()` to force SvelteKit to re-run load functions with fresh session data.

### Load Functions for Param Extraction

Use `+page.ts`/`+layout.ts` load functions to extract route params and query strings.
