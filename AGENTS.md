# Chatto — Agent Guidelines

Authoritative guidance for coding agents working in the Chatto codebase.
Use this file as the primary source for project-specific conventions, architectural constraints, and implementation gotchas.

## How to Use This File

- Read the section relevant to the area you are changing.
- For backend architecture, streams, KV buckets, and data flow, also read `docs/ARCHITECTURE.md`.
- For architectural rationale and tradeoffs, read `docs/adr/`.
- Keep tests and documentation in sync with code changes.

## Section Scope

- **Project Overview / Project Status / General Guidelines / Git & CI** — repository-wide
- **Backend Development** — primarily `cli/**`
- **Frontend Development** — primarily `frontend/**`
- **GraphQL Development** — `**/*.graphqls`, `cli/internal/graph/**`, `frontend/src/lib/graphql/**`
- **Authorization & Admin** — auth-sensitive backend, GraphQL, and frontend code
- **Backup & Restore** — primarily `cli/cmd/backup.go`, `cli/cmd/restore.go`, `cli/cmd/keys.go`
- **Documentation Website** — `docs-website/**`

## Source-of-Truth Notes

- `docs/ARCHITECTURE.md` is the source of truth for backend architecture, streams, KV buckets, and data flow.
- `docs/adr/` explains architectural decisions and tradeoffs.
- This file summarizes project conventions and implementation rules for coding agents.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Project Status](#project-status)
3. [General Guidelines](#general-guidelines)
4. [Git & CI](#git--ci)
5. [Backend Development](#backend-development)
6. [Frontend Development](#frontend-development)
7. [GraphQL Development](#graphql-development)
8. [Authorization & Admin](#authorization--admin)
9. [Backup & Restore](#backup--restore)
10. [Documentation Website](#documentation-website)

---

## Project Overview

Chatto is a real-time chat application with a GraphQL gateway backed by NATS/JetStream. Single Go executable with embedded NATS server + web server; SvelteKit frontend connects via GraphQL queries/mutations/subscriptions.

### Key Files

- `docs/ARCHITECTURE.md` - Source of truth for architecture (streams, KV buckets, data flow)
- `docs/adr/` - Architecture Decision Records explaining *why* key decisions were made
- `.beans/` - Issue tracking (use `beans` CLI to manage tasks and roadmap)

### Monorepo Structure

- `cli/` - Go backend (Gin, gqlgen, embedded NATS)
- `frontend/` - SvelteKit SPA (Svelte 5 runes, urql, TailwindCSS v4)
- `proto/` - Protocol Buffer definitions
- `examples/` - Various example configurations for self-hosting Chatto

### Basic Architecture

- NATS for pubsub and persistent storage (JetStream)
- "ChattoCore", a Go package with low-level domain logic that interacts with NATS. Receives information on the actor, but does not perform authorization.
- GraphQL gateway (gqlgen) that calls ChattoCore directly, performing both authentication and authorization before calling Core methods.

### Running Tools

**IMPORTANT:** Tools like `go`, `pnpm`, `node`, etc. are managed by mise and are NOT directly available in PATH. You must use `mise x --` to run them:

```bash
# Correct - use mise x to run tools
mise x -- pnpm install
mise x -- go test ./...
mise x -- node script.js

# Incorrect - tools are not in PATH
pnpm install        # Will fail
go test ./...       # Will fail
```

Prefer using mise tasks (below) when available, as they handle this automatically.

### Commands

All commands via mise task runner. Run from project root.

| Command                 | Description                                   |
| ----------------------- | --------------------------------------------- |
| `mise dev`              | Run full dev environment (backend + frontend) |
| `mise test`             | Run all tests                                 |
| `mise test-cli`         | Run Go tests only                             |
| `mise test-frontend`    | Run frontend tests only                       |
| `mise build`            | Build for all architectures                   |
| `mise codegen`          | Run all code generation                       |
| `mise codegen-cli`      | Generate Go code (GraphQL + proto)            |
| `mise codegen-frontend` | Generate frontend GraphQL types               |
| `mise clean`            | Remove build artifacts                        |
| `mise docs`             | Start pkgsite documentation server            |
| `mise bump`             | Bump version (uses 'goversion' tool)          |

### Reference Links

- [NATS Go SDK](https://pkg.go.dev/github.com/nats-io/nats.go)
- [JetStream Go SDK](https://pkg.go.dev/github.com/nats-io/nats.go/jetstream)
- [gqlgen](https://gqlgen.com/)

---

## Project Status

### Project Maturity

Chatto has public instances running with real user data. Data migrations are required for any breaking changes to storage schemas or APIs. Never assume data can be discarded.

### Breaking vs. Non-Breaking Changes

- Any changes to the `proto/` definitions are considered **SIGNIFICANT BREAKING CHANGES**. These affect compatibility with existing clients and servers, and require careful coordination for deployment.
- Any changes to the GraphQL schema files in the `cli/internal/graph/` package should be considered **POTENTIALLY BREAKING CHANGES**. GraphQL is our primary API; at this point in the project we don't have external clients, though, but we should still tread carefully.
- Any changes to **NATS JetStream stream/KV schemas** (stream names, subject patterns, KV key formats) are **BREAKING CHANGES** that require a data migration path. Existing instances must be migrated on upgrade without data loss.
- All other changes can be considered non-breaking.

---

## General Guidelines

### Prime Directives

- DO keep tests and documentation up to date when making changes (see [Documentation Website](#documentation-website) for docs website specifics)
- DON'T create commits unless explicitly asked to do so

### Priorities

- Efficient resource usage (CPU, memory, storage) to minimize hosting costs
- Single executable with everything out of the box (exceptions for pluggable advanced features like full-text search)
- Reliable message sending and delivery is OF THE HIGHEST PRIORITY - must be ROCK-SOLID

### Code Style & Approach

- **Prefer the simplest possible approach first.** Do not over-engineer solutions. If a fix can be done in 2-3 lines, do not create abstractions, wrappers, or complex architectures. Wait for user feedback before adding complexity.
- **When fixing bugs involving caches or state, prefer minimal, targeted invalidation** over clearing entire caches. Avoid full-page reload flashes or broad cache wipes. Only invalidate the specific stale data.
- **Functions that depend on "which instance" should require an explicit instance ID parameter.** Don't default to a global "current instance" — it creates coupling and timing bugs. Navigation helpers, storage functions, and state lookups all take `instanceId` as the first parameter.

### UI & Frontend Style

- **Don't default to smaller font sizes.** Use the base text size unless there's a clear reason to go smaller (e.g., timestamps, metadata footnotes). Only use `text-xs` or `text-sm` when explicitly asked or when it's an established pattern in the codebase for that specific element type.
- **Keep text and labels aligned.** When elements appear above or below each other in a layout, they should share the same indentation. Use the same flex layout structure (spacers, gaps, padding) rather than calculating offsets manually. If a label sits above a content block, mirror the content block's layout so they line up naturally.
- **Always add `cursor-pointer` to clickable elements.** Buttons, toggles, and other interactive elements must have `cursor-pointer` in their class list. Tailwind CSS v4 does not add this automatically for buttons.
- **Never use `{@html}`.** It bypasses Svelte's XSS protection. Use snippets or components to compose rich content instead. Even for "safe" hardcoded strings — it sets a bad precedent and makes auditing harder.
- **Use `<SkeletonImg>` instead of `<img class="skeleton">`.** The `.skeleton` CSS utility adds a shimmer background that looks wrong behind transparent PNGs and should only show while loading. Use the `SkeletonImg` component (`$lib/ui/SkeletonImg.svelte`) which reactively applies `.skeleton` until `onload` fires. Never use imperative `classList.remove()` in a reactive framework — track state declaratively instead.

### Planning

- When planning features with beans, separate frontend and backend into distinct tasks
- Each bean should be one small, focused PR that can be reviewed and merged independently
- **Gather broader context before finalizing scope** - Review related files and patterns across the codebase before committing to a task's boundaries
- **Start with most impactful/dependent tasks** - If task A is a prerequisite for task B, complete A first to unblock further work
- **Research before implementing** - Before suggesting optimizations or implementation approaches for NATS/JetStream, Go embed, or other infrastructure tools, read the actual documentation and SDK source first. Do not guess at API capabilities.

### iOS Safari Gotchas

- **`pointer-events: none` doesn't block touch-scroll.** It only suppresses click/tap events. iOS Safari routes touch-scroll gestures at the compositor level, bypassing `pointer-events`. When overlaying scrollable content (e.g., a slideover pane over a scrollable list), use the HTML `inert` attribute instead — it truly disables all interaction including scroll. (Safari 15.5+, Chrome 102+, Firefox 112+.)
- **Use `overscroll-y-contain` on isolated scroll containers.** Prevents scroll chaining — where scrolling past the edge of one container starts scrolling a parent or sibling. Good default for chat message lists and any panel that shouldn't leak scroll gestures.

### E2E Test Anti-Flakiness

- **Never use raw `waitForTimeout(number)`.** Always use a `TIMEOUTS.*` constant from `e2e/constants.ts`. This makes global tuning a single-line change and makes the intent clear.
- **Prefer observable state over fixed delays.** Instead of `waitForTimeout(500)`, use `expect(locator).toBeVisible()` or `toPass()` with polling intervals. Fixed delays don't adapt to CI slowness.
- **Use `toPass()` with exponential backoff for negative assertions.** When asserting something should NOT happen (e.g., no unread dot after a thread reply), use `toPass()` with `POLLING_INTERVALS` to give events time to propagate before asserting absence.
- **Use `waitForLoadState('networkidle')` for hydration waits.** Instead of guessing how long SvelteKit hydration takes with a fixed delay, wait for network activity to settle.
- **Scroll settling needs `TIMEOUTS.SCROLL_SETTLE` (150ms).** Between wheel events, virtua needs time to process measurements and scroll corrections. 50ms is too tight for CI; use the constant.
- **The only acceptable raw delay is for wall-clock timing tests** (e.g., cookie timestamp comparison that needs >1s to pass). These must have a comment explaining why.

### Refactoring

- **Keep refactoring PRs small and focused** - Don't let scope creep. If a refactor reveals additional cleanup opportunities, create separate beans for them rather than bundling everything together.
- **Verify regressions before fixing** - When a bug is reported during a refactor, first write a failing test that reproduces it, then investigate. Don't dive into code archaeology without a reproducible case.
- **Avoid over-engineering early** - Evaluate whether a complex abstraction provides sufficient value. If each use case has unique logic that differs significantly, simpler focused components may be better than a highly configurable wrapper.

---

## Git & CI

### Making Git Commits

- Use conventional commits with a clear, descriptive message.
- Use the commit description for a bullet list of changes made.

### Creating Pull Requests

- After creating a PR, always check that CI passes. If CI fails, proactively diagnose and fix the failures without waiting to be asked.
- **The baseline for test failures is ALWAYS `main`, never the previous commit on the branch.** If a test passes on `main` but fails on your branch, it is a regression you introduced and you MUST fix it. Do not dismiss a failure just because a previous commit on the same branch also had it. The only tests you may ignore are those that are also failing or flaky on `main`.
- Common CI failure sources: broken tests from removed code paths, nil loggers in test setup, ESLint missing keys in Svelte `{#each}` blocks, and test selectors that are too broad.

### Merging Pull Requests

- Before merging a PR, first merge `origin/main` into the branch to ensure it's up-to-date.
- Run tests after merging to catch any integration issues before the final merge.

---

## Backend Development

> Applies primarily to files in `cli/`.

### Architecture

- `ChattoCore` handles all domain operations (spaces, users, rooms, messages)
- KV buckets are source of truth; event streams provide audit trail
- IDs: 14-char NanoID via `helpers.NewID()` (~83.4 bits entropy)
- When adding streams/KV buckets, update `docs/ARCHITECTURE.md` inventory

### NATS/JetStream Tips

- `kv.Create()` for atomic insert (fails if key exists) vs `kv.Put()` for upsert
- `kv.ListKeysFiltered(ctx, ...filters)` for efficient key queries
- Consumer names can't have dots; KV keys use dots (act as NATS subjects)
- Streams/KV are immediately consistent within the cluster
- KV keys can't contain arbitrary Unicode; use named identifiers (e.g., `thumbsup` not `thumbs-up-emoji`)
- **Optimistic locking**: When updating KV entries that may have concurrent writers, use revision-based updates:
  1. Get current entry with `kv.Get()` to obtain its revision
  2. Use `kv.Update(ctx, key, value, revision)` for atomic update (fails if revision changed)
  3. For new keys, use `kv.Create()` instead (fails if key exists)
  4. Retry on `jetstream.ErrKeyExists` up to a max attempts (e.g., 5 retries)
- **Subject structure changes are high-risk**: Changes to NATS subject patterns cascade into stream configs, consumer filters, and query logic (e.g., `GetLastMsgForSubject`, `WithSubjectFilter`). They need careful end-to-end verification including e2e tests.
- **Space streams are lazily initialized**: `getSpaceStream()` calls `ensureSpaceStream()` on first access, using a `sync.Map` cache to avoid redundant `CreateOrUpdateStream` calls. This means stream config changes are applied on-demand rather than at startup.

### Room Event Query Behavior

`GetRoomEvents` uses three optimized code paths based on room size and query type:

- **Small room fast path**: Uses `stream.Info(WithSubjectFilter)` to check total event count. If <= `limit`, fetches everything in one consumer with `DeliverAllPolicy`
- **Initial load (large rooms)**: Uses `GetLastMsgForSubject` (O(1) lookup) to find the room's last event sequence, then starts a consumer near the end using `DeliverByStartSequencePolicy` with progressive multipliers (3x, 10x, 50x). Falls back to `DeliverAllPolicy` if needed
- **Pagination (large rooms)**: Uses `beforeTime` as the cursor. Tries a single 30-day window, falling back to `DeliverAllPolicy` if insufficient events are found

**Important**: `room_last_msg_at` in the RUNTIME bucket only tracks MESSAGE timestamps, not join/leave events. This is intentional for sorting by "recent activity" (conversations with recent messages). Don't rely on this field to determine when the most recent _event_ occurred.

### Event Patterns

- **JetStream events**: For data needing audit trail, ordering, or replay (messages, memberships)
- **Live-only events**: For transient UI updates where KV is source of truth (reactions, typing indicators)
  - Publish to `live.` subject prefix via `publishLiveEvent()`
  - Bypasses JetStream storage entirely
- **Adding new live event types** requires updates in TWO places:
  1. Core handler: Add case in `StreamMySpaceLiveEvents` (`core.go`)
  2. GraphQL resolver: Add case in `liveEventResolver.Event` (`events.resolvers.go`)
  - Missing the resolver case causes events to silently fail at the GraphQL layer
- **Avoid fan-out on publish**: When broadcasting events to multiple users (e.g., space updates), do NOT iterate through recipients and publish to each. Instead:
  - Publish once to a scoped subject (e.g., `instance.space.{spaceId}.updated`)
  - Use server-side authorization filtering in the subscription handler
  - This scales to large numbers of users without N publish operations

### Instance Event Authorization

Instance events use subject pattern `live.instance.{scope}.{id}.{eventType}` and are filtered by `isAuthorizedForInstanceEvent` in `core.go`:

| Scope   | Subject Pattern                   | Delivered To                    |
| ------- | --------------------------------- | ------------------------------- |
| `user`  | `live.instance.user.{userId}.*`   | Only that user (private events) |
| `space` | `live.instance.space.{spaceId}.*` | All space members               |

**Adding a new instance event type:**

1. Add protobuf message to `live_event.proto`
2. Add to GraphQL schema in `events.graphqls` (type + union)
3. Add `IsInstanceEventType()` method in `pb/chatto/core/v1/graphql.go`
4. Add case in `unwrapInstanceEvent()` in `event_helpers.go`
5. Publish using `subjects.InstanceUserEvent()` or `subjects.InstanceSpaceEvent()`
6. Subscribe in frontend via `instanceEventBus.svelte.ts`

**When to create a live event:** Any time a user action changes state that other tabs/devices or other UI components need to reflect in real-time. Common triggers:
- User changes a preference or setting (notification level, follow state)
- Server-side auto-mutations (auto-follow on posting to a thread)
- Cross-tab sync needs (reading a room in one tab should update indicators in others)

If a mutation changes state visible in the UI and you don't publish a live event, the UI will be stale until refresh. Always consider: "Will other tabs or other components on the same page need to know about this change?"

**Broadcasting user events to everyone**: By default, user-scoped events are private (only delivered to that user). To broadcast an event to all authenticated users (e.g., profile updates since profiles are public), add an explicit check in `isAuthorizedForInstanceEvent`:

```go
case "user":
    if eventType == "profile_updated" {
        return true  // Broadcast to all
    }
    return scopeID == userID  // Private to user
```

### NATS Subject Patterns

#### Design Principles

When designing NATS subject patterns, follow these principles:

**1. Unified Namespaces for Related Events**

Group related events under a common prefix so a single wildcard subscription captures all of them:

```
# Good: All messages (root + thread) under msg.>
space.{s}.room.{r}.msg.{eventId}                    # Root message
space.{s}.room.{r}.msg.{rootId}.replies.{eventId}   # Thread reply

# Bad: Separate namespaces require multiple subscriptions
space.{s}.room.{r}.msg.{eventId}                    # Root message
space.{s}.room.{r}.thread.{rootId}.{eventId}        # Thread reply
```

**2. Semantic Markers for Disambiguation**

Use explicit semantic tokens (like `.replies.`) to distinguish subject types, rather than relying on part counts alone:

```
# Good: Clear semantic marker
msg.{rootId}.replies.{eventId}   # "replies" explicitly marks thread messages

# Less clear: Only part count differs
msg.{eventId}                    # Root (6 parts)
thread.{rootId}.{eventId}        # Thread (7 parts)
```

**3. Hierarchical Nesting**

Structure subjects so children nest under parents in the namespace:

```
# Good: Threads nest under their root message
msg.{rootId}.replies.{eventId}

# Less intuitive: Separate top-level namespace
thread.{rootId}.{eventId}
```

#### Filtering Patterns Reference

For room messages, these wildcard patterns enable efficient filtering:

| Pattern | Matches |
|---------|---------|
| `msg.>` | All messages (root + threads) |
| `msg.*` | Root messages only |
| `msg.*.replies.>` | All thread replies (any thread) |
| `msg.{rootId}.replies.>` | Replies in a specific thread |
| `msg.*.replies.{eventId}` | Lookup thread reply by event ID |

#### Subject Refactoring Checklist

When changing subject patterns:

1. **Update construction functions** in `subjects.go` (e.g., `SpaceRoomThread`)
2. **Update parsing functions** in `subjects.go` (e.g., `IsThreadSubject`, `ParseEventIDFromSubject`)
3. **Update all test expectations** in `subjects_test.go`
4. **Update comments** in files that reference the patterns (e.g., `rooms.go`)
5. **Update `docs/ARCHITECTURE.md`** subject tables and filtering examples
6. **Run full test suite** including e2e tests - subject changes cascade through the entire system

Subject changes are high-risk because they affect:
- JetStream stream configs and filters
- Consumer subscriptions
- `GetLastMsgForSubject` lookups
- Event routing and delivery

### Adding New Live Space Event Types

When adding a new live event type (published via `publishLiveSpaceEvent`), you must update **all** of these locations or the event will be silently dropped:

#### Backend Checklist

1. **Proto definition** — Add the event message to `proto/chatto/core/v1/live_event.proto` and the oneof case to `space_event.proto`
2. **Proto interface method** — Add `IsSpaceEventType()` to `cli/internal/pb/chatto/core/v1/graphql.go` so gqlgen recognizes the proto type
3. **Event unwrapper** — Add a case in `unwrapSpaceEvent()` in `cli/internal/graph/event_helpers.go`
4. **Subscription room ID extraction** — Add a case to the `liveRoomMsgChan` switch in `StreamMySpaceEvents` (`cli/internal/core/core.go` ~line 1758) that extracts the `roomID` from the event. **Without this, the event is silently dropped** because `roomID` stays empty and the `if roomID == ""` guard skips it.
5. **GraphQL schema** — Add the type definition and include it in the `SpaceEventType` union in `events.graphqls`

#### Frontend Checklist

6. **GraphQL fragment** — Add the event fields to the subscription fragment in `RoomEvent.svelte`
7. **Event handler** — Handle the event in `RoomEventsPane.svelte` (and `ThreadPane.svelte` if applicable)

#### Common Pitfall: Room ID Extraction

The subscription handler in `core.go` routes live room events by extracting the `roomID` from each event type via a type switch. If your new event type isn't in that switch, the room ID will be empty and the event will be silently dropped with `continue`. There is no error log — it just disappears. Always check this switch when adding live events.

#### Common Pitfall: Race Conditions with KV State

If your event relies on KV state that's set alongside the event (e.g., setting a processing status in RUNTIME KV and then publishing an event), make sure the KV write happens **before** the action that triggers the subscription event. The subscription delivers events immediately, and field resolvers that read KV will see stale/missing data if the write hasn't happened yet.

Example: Video processing sets PENDING state in KV *before* `PostMessage` publishes to JetStream, so that when the subscription resolves `Attachment.videoProcessing`, the KV entry already exists.

### Image Processing

- **nativewebp is lossless only**: `github.com/HugoSmits86/nativewebp` encodes VP8L (lossless WebP). There is no lossy quality option — the `Options` struct only has `UseExtendedFormat` for metadata containers. If lossy WebP is needed in the future, a different library would be required.
- **Thumbnail encoding is format-aware**: `TransformImage()` picks the output format based on the input:
  - **Animated GIF** -> WebP (lossless, with proper frame compositing and disposal handling)
  - **Transparent static** -> WebP (lossless, preserves alpha)
  - **Opaque static** -> JPEG (lossy q80, smaller files)

  Opaque static images use JPEG rather than WebP because nativewebp is lossless-only, which would produce larger files for photos.
- **Image cache stores raw bytes without format metadata**. Use `DetectImageContentType()` (magic bytes) when serving cached images — never hardcode a content type.

### Service Lifecycle

- Long-running services use `Run(ctx context.Context) error` — blocks until ctx cancelled
- Use `signal.NotifyContext` for shutdown signals (not manual goroutine + channel)
- Use `errgroup` to coordinate multiple concurrent blocking services

### API Design

- Use GraphQL for all client-facing APIs - avoid REST endpoints for application logic
- gqlgen supports file uploads via the `Upload` scalar ([docs](https://gqlgen.com/reference/file-upload/))
- REST endpoints are acceptable only for: OAuth callbacks, webhooks, health checks, and pre-auth discovery (e.g., `GET /api/instance` for multi-instance client probing before GraphQL setup)

### Dataloaders

- Dataloaders are injected for **HTTP requests only**, not WebSocket connections
- WebSocket connections are long-lived; dataloader caches would become stale across subscription events (e.g., user updates display name mid-session)
- Subscription resolvers fall back to direct `core.Get*()` calls via helper methods like `r.getUser()`
- This is intentional: HTTP requests benefit from batching (loading room history with many reactions), while subscription events arrive one at a time and don't benefit from batching anyway

### Security

- All GraphQL mutations must check permissions via `core.RequirePermission()`

### Known Test Issues

- `TestAuthRoutes_TestEmailEndpoint` in `cli/internal/http_server/` is a pre-existing failure — do not investigate as a regression.

### Cost Reference

Hetzner volumes: ~53 EUR/TB with R3 replication (3x storage)

---

## Frontend Development

> Applies primarily to files in `frontend/`.

### Svelte 5 Lifecycle Timing

`$effect` runs AFTER the initial render pass completes. Component script initialization (top-level `<script>` code) runs synchronously DURING render. This means:

- Parent script init -> Parent template renders -> **Child script init** -> Child template renders -> `onMount` (children first) -> `$effect`

**Gotcha:** If a parent's `$effect` creates a resource (e.g., an event bus), child components cannot access it during their script initialization — it doesn't exist yet.

**Fix:** Create resources synchronously in the parent's script section for the initial render. Use `$effect` only for reactive changes after mount (additions/removals). See `+layout.svelte` bus initialization pattern:

```svelte
// Synchronous: available to children during render
for (const instance of instanceRegistry.instances) {
  instanceEventBusManager.startBus(instance.id, client);
}

// Reactive: handles additions/removals after initial render
$effect(() => {
  // startBus is idempotent — no-op if already started
  for (const instance of instanceRegistry.instances) {
    if (!instanceEventBusManager.getBus(instance.id)) {
      instanceEventBusManager.startBus(instance.id, client);
    }
  }
  return () => { /* cleanup */ };
});
```

### Context Getters Must Be Wrapped in `$derived`

When reading a context value that depends on async data (e.g., `getRoomPermissions()`, `getRoomMembers()`), **always wrap the call in `$derived`**. A plain `const` snapshots the value at script init time — if the underlying data hasn't loaded yet, you get the default/empty value permanently.

```ts
// BAD: snapshots DEFAULT_PERMISSIONS during init, never updates
const roomPermissions = getRoomPermissions();

// GOOD: re-evaluates reactively when the underlying data loads
const roomPermissions = $derived(getRoomPermissions());
```

This applies to any `getXxx()` context getter backed by a `$derived` or `$state` in a parent component. The getter chain only propagates reactivity when read inside a reactive context (`$derived`, `$effect`, or template expression).

### Pass Reactive Props as Getters to Non-Reactive Init Functions

When a `use*` hook or init function needs a prop value but runs during script init (non-reactive), **pass the prop as a getter** so the read happens inside `$effect` rather than at the call site. Never suppress `state_referenced_locally` warnings — they indicate a real reactivity bug.

```ts
// BAD: reads data.user during script init — not reactive, triggers warning
useInstanceRegistry(data.user);

// BAD: suppressing the warning hides the bug
// svelte-ignore state_referenced_locally
useInstanceRegistry(data.user);

// GOOD: getter defers the read to a reactive context ($effect)
useInstanceRegistry(() => data.user);
```

Inside the hook, call the getter within `$effect` for reactivity, and optionally once synchronously if downstream init code needs the value immediately (only if the called function is idempotent):

```ts
export function useInstanceRegistry(getUser: () => unknown): void {
  doSomething(!!getUser());        // Sync: immediate availability
  $effect(() => {
    doSomething(!!getUser());      // Reactive: responds to changes
  });
}
```

### Three-State Guards and `effect_update_depth_exceeded`

When using the three-state pattern (`undefined` = loading, `null` = not found, `object` = loaded) with async data hooks like `useRoomData`, the template guard must block rendering during loading:

```svelte
<!-- BAD: undefined !== null is true — children mount during loading.
     When data loads, the cascade of derived values and effects can exceed
     Svelte 5's effect_update_depth_exceeded limit. -->
{#if room.roomData !== null}

<!-- GOOD: blocks rendering for both undefined (loading) and null (not found) -->
{#if room.roomData}
```

The cascade happens because all children mount with default/empty state, then when data arrives, every derived value, context getter, and effect updates simultaneously. With enough children (composer, event list, permissions, members, etc.), this exceeds Svelte's depth limit and silently halts state propagation.

### Module-Level State Must Use `<script module>`

State that must survive component unmount/remount cycles (e.g., draft file stash maps, singleton caches) **must** be in `<script module>` — not in `<script>`. Instance-level `<script>` code runs on every mount, creating fresh state each time.

```svelte
<!-- Module-level: persists across component instances -->
<script module lang="ts">
  const draftFilesMap = new Map<string, FileWithUrl[]>();
</script>

<script lang="ts">
  // BAD: this creates a NEW map on every mount — stashed data is lost
  // const draftFilesMap = new Map<string, FileWithUrl[]>();
</script>
```

This matters when parent guards (`{#if data}`) cause child components to unmount during loading transitions and remount when data arrives. On the previous architecture where parent data was held via `$derived(await ...)`, components stayed mounted across transitions, hiding this class of bug.

### Multi-Instance Architecture

The frontend supports connecting to multiple Chatto instances simultaneously. The `InstanceRegistry` (singleton at `instanceRegistry`) owns both registration data and per-instance state stores, ensuring atomic creation — no race between "instance registered" and "store exists."

#### URL is the Source of Truth

The URL determines which instance is active:

- Landing page: `/` (welcome or redirect to Browse Spaces)
- Origin instance: `/chat/-/SAbcDef/RAbcDef` (`-` = the server hosting the SPA)
- Remote instance: `/chat/chat.hmans.dev/SAbcDef/RAbcDef` (hostname)

The `[instanceId]/+layout.svelte` resolves the URL segment to an instance ID via `segmentToInstanceId()` and provides it via Svelte context. Components inside this route tree use:

- `graphqlClientManager.getClient(instanceId)` for the correct GraphQL client
- `instanceRegistry.getStore(instanceId)` for per-instance state (notifications, permissions, etc.)
- `resolve()` from `$app/paths` for navigation (see Navigation section below)

**Key rule:** Never store "which instance is active" in runtime state. Always derive it from the URL.

#### No "Home Instance" Flag

There is no `isHome` flag. The origin instance (the server serving the SPA) is detected by comparing `instance.url` against `window.location.origin`:

- `instanceRegistry.originInstance` — finds the matching instance (or `undefined`)
- `instanceRegistry.isOriginInstance(id)` — checks a specific instance

The origin uses cookie auth (`token: null`). Remote instances use bearer tokens (`token: string`).

#### Origin Auto-Registration

On app init, `probeOrigin()` detects whether the SPA is served by a Chatto instance by fetching `/api/instance`. If it responds, the origin is auto-registered. If it fails (static hosting), nothing happens. This is idempotent — no-ops if already registered.

#### Per-Instance Permissions

Each `InstanceStateStore` has a `permissions` field (`InstancePermissions`) loaded by `InstanceSpaceSection` from the `viewer` query. Use `instanceRegistry.getStore(id).permissions` to check per-instance capabilities (e.g., `canCreateSpace`).

#### Disconnect vs Sign Out

- **Disconnect** (`removeInstance`): Removes one instance. Cleans up event bus, store, and GraphQL client. If it's the origin, also revokes the cookie session and hard-reloads.
- **Sign Out** (`removeAll`): Removes ALL instances, revokes origin cookie, hard-reloads to `/`.

#### CORS Boundary

`/api/instance` (REST) has wildcard CORS (`Access-Control-Allow-Origin: *`) — it's the **only** endpoint usable cross-origin without configuration. `/api/graphql` requires the client's origin in the allowed list. The add-instance flow must use `/api/instance` for probing remote instances. Rich instance data (welcome message, config) should be included in `/api/instance` when needed cross-origin, since GraphQL isn't accessible pre-registration.

### Use `createContext` for Svelte Context

Always use the `createContext` API from Svelte instead of manual `Symbol` keys with `getContext`/`setContext`. Re-export the `[get, set]` tuple directly — don't wrap them in functions unless the wrapper adds real logic (e.g., constructing the value, transforming the return):

```ts
import { createContext } from 'svelte';

// GOOD: re-export directly
export const [getMyContext, setMyContext] = createContext<MyType>();

// GOOD: factory wrapper that adds real logic (construction, options)
export const [getComposerContext, setComposerContext] = createContext<ComposerContext>();
export function createComposerContext(options?: Options): ComposerContext {
  const ctx = new ComposerContext(options);
  setComposerContext(ctx);
  return ctx;
}

// BAD: manual Symbol keys
const KEY = Symbol('myContext');
setContext(KEY, value);
getContext<MyType>(KEY);

// BAD: pointless wrapper around get/set
const [getCtx, setCtx] = createContext<MyType>();
export function getMyContext() { return getCtx(); }  // just re-export instead
```

### Prefer Context Over Mutable Singletons for URL-Derived State

When state is derived from the URL (route params), provide it via Svelte context from a layout — don't sync it into a mutable singleton via `$effect`.

```ts
// BAD: syncing URL param into mutable state creates timing bugs
$effect.pre(() => {
  registry.activeId = resolveFromUrl(page.params.instanceId);
});
// Children may render before the effect runs, seeing stale state

// GOOD: context is available synchronously during child script init
const instanceId = $derived(resolveFromUrl(page.params.instanceId));
setActiveInstance(() => instanceId);
```

The `$effect`/`$effect.pre` approach fails because effects run after render — child components initialize with the old value and only see the update on the next tick. Context set during script initialization is available to children immediately.

### Navigation

**Always use `resolve()` from `$app/paths` for all internal navigation** — both `href` attributes and `goto()` calls. Use route IDs with params, not manually constructed path strings:

```ts
import { resolve } from '$app/paths';
import { instanceIdToSegment } from '$lib/navigation';

// GOOD: type-safe route ID with params
resolve('/chat/[instanceId]/[spaceId]/[roomId]', { instanceId: instanceSegment, spaceId, roomId })

// BAD: manual string construction
`/chat/${instanceSegment}/${spaceId}/${roomId}`
```

`$lib/navigation` only exports two conversion functions: `instanceIdToSegment(id)` (instance ID -> URL segment) and `segmentToInstanceId(segment)` (URL segment -> instance ID). There are no path builder helpers — use `resolve()` directly.

**DRY tip:** In components with 3+ resolve calls using the same instance, derive the segment once:

```ts
const getInstanceId = getActiveInstance();
const instanceSegment = $derived(instanceIdToSegment(getInstanceId()));
```

**DM routes use `[instanceSegment]`**, not `[instanceId]`:

```ts
resolve('/chat/dm/[instanceSegment]/[conversationId]', { instanceSegment, conversationId })
```

---

## GraphQL Development

> Applies primarily to `**/*.graphqls`, `cli/internal/graph/**`, and `frontend/src/lib/graphql/**`.

gqlgen is schema-first. Follow this workflow for GraphQL changes:

1. Update schema files (`*.graphqls`) first
2. Run `mise codegen-cli` to regenerate Go types/resolvers
3. After frontend query changes, run `mise codegen-frontend`

### Schema Documentation

- Every type must have a type-level description
- Every field must have a description, even "obvious" relationship fields
- Every enum value must have a description explaining its meaning
- Descriptions should be concise (one sentence preferred)
- Include format examples for non-obvious string values (e.g., `"Round-trip time (e.g., '1.234ms')."`)

### Schema Directives

Use gqlgen directives to control code generation:

#### `@goField(forceResolver: true)`

Add this to fields that have custom resolvers. This:

- Silences "adding resolver method for X, nothing matched" warnings
- Documents that the field requires a resolver (not auto-bound from protobuf)
- Required for fields that are computed, lazy-loaded, or need authorization

```graphql
type Space {
  id: ID! # Auto-bound from protobuf
  rooms: [Room!]! @goField(forceResolver: true) # Requires resolver
  viewerIsMember: Boolean! @goField(forceResolver: true) # Computed field
}
```

#### `@goModel(model: "...")`

Bind a GraphQL type to a specific Go type:

```graphql
scalar Time @goModel(model: "hmans.de/chatto/internal/graph.Time")
```

### Unions vs Interfaces

**Prefer unions over interfaces** for polymorphic types. This project uses unions (like `EventType`, `NotificationItem`) rather than interfaces:

- **Union**: Simpler Go models - only need `IsTypeName()` marker method
- **Interface**: Requires getter methods for all shared fields in Go

With unions, clients check `__typename` and use inline fragments to query fields:

```graphql
# Union requires inline fragments for ALL fields (including shared ones)
query {
  notifications {
    __typename
    ... on DMMessageNotificationItem {
      id
      createdAt
      actor {
        id
      }
      summary
      room {
        id
      }
    }
    ... on MentionNotificationItem {
      id
      createdAt
      actor {
        id
      }
      summary
      space {
        id
      }
      room {
        id
      }
    }
  }
}
```

### Custom Model Files

When gqlgen's auto-binding doesn't work (e.g., types needing internal fields for resolvers), create custom models in `cli/internal/graph/model/`. Keep these minimal:

```go
type MyCustomType struct {
    ID      string `json:"id"`
    // Internal fields for resolvers (not exposed in GraphQL)
    ActorID string `json:"-"`
}

func (MyCustomType) IsMyUnion() {}  // Union marker method
```

### Type Compatibility

When autobind can't match protobuf types to GraphQL types, you'll see warnings like:

- `Time is incompatible with *timestamppb.Timestamp` -> Use custom scalar with `@goModel`
- `ID is incompatible with uint64` -> Add resolver to convert types
- `adding resolver method for X, nothing matched` -> Add `@goField(forceResolver: true)`

### Optimistic UI vs Backend Enforcement

For permission-based UI gating (e.g., "can viewer manage this user?"):

- **Frontend handles optimistic checks** using locally available data (role positions, membership status)
- **Backend enforces authorization** on mutations - the actual security boundary

Don't add `viewer*` boolean fields that require backend round-trips when the frontend already has the data to compute them. Instead:

| Avoid                                    | Prefer                                 |
| ---------------------------------------- | -------------------------------------- |
| `SpaceMember.viewerCanManage: Boolean!`  | Frontend computes using role positions |
| Fetching permissions for every list item | Query permissions once, apply locally  |

Backend `viewer*` fields are still useful for:

- Complex authorization logic the frontend can't replicate
- Fallback when local data isn't available
- API-level authorization (e.g., `Space.viewerCanManageUser(userId)` for mutations)

### Prefer Core Types Over Wrapper Types

Avoid creating wrapper types that just add fields to existing types. Instead, add scoped fields to the core type:

| Avoid                                            | Prefer                                            |
| ------------------------------------------------ | ------------------------------------------------- |
| `SpaceMember { user: User!, roles: [String!]! }` | `User.spaceRoles(spaceId: ID!): [String!]!`       |
| `Space.members: [SpaceMember!]!`                 | `Space.members: [User!]!` with `spaceRoles` field |

Benefits:

- Simpler schema with fewer types
- Consistent data shape - a User is always a User
- Easier caching and normalization in clients
- Fields can be queried in any context where the User is available

### Fragment Type Assertions

When using gql.tada/graphql-codegen with fragments, the generated TypeScript types use `$fragmentRefs` markers that don't expose fields like `id` directly. The fields exist at runtime, but TypeScript requires assertions:

```typescript
// TypeScript error: Property 'id' does not exist on type '{ __typename?: "User" } & { $fragmentRefs?: ... }'
const actorId = participant.id;

// Works - field exists at runtime, assertion satisfies TypeScript
const actorId = (participant as { id?: string })?.id;
```

This commonly occurs when working with fragment wrapper types from subscriptions or when comparing actors/participants by ID.

---

## Authorization & Admin

### Core Principles

1. **Users are bound to an instance** - All users exist within a single Chatto instance
2. **Spaces are discoverable** - Users can browse all spaces for discovery purposes
3. **Room access requires space membership** - Users must join a space before accessing its rooms
4. **Message access requires room membership** - Users can only read/write messages in rooms they've joined
5. **User profiles are public** - Basic user info (id, login, displayName, avatar) is visible to all authenticated users
6. **Membership info is private** - Users can only see their own space/room memberships

### Authorization Architecture

Authorization is enforced at the **API boundary**, not in core:

| Layer | Responsibility |
|-------|----------------|
| **GraphQL** | User-facing API. Checks authorization via `Can*` functions before calling core. |
| **Core** | Pure business logic. Assumes caller is authorized. Documents requirements in comments. |
| **NATS** | Extension/internal API. Trusted context, calls core directly. |

**Why this design:**
- Core functions are reusable from trusted contexts (NATS handlers, background jobs)
- No redundant permission checks when core calls other core functions
- Clear separation: GraphQL handles user authorization, core handles business logic
- Audit logging can be added orthogonally without coupling to authorization

### Permission System

Permissions are granted through roles assigned to space members. Use `Can*` functions in `core/can.go` to check permissions.

#### Hierarchy-Wins Resolution

Permission resolution follows role hierarchy order (lower position = higher rank):

1. Get user's roles sorted by position (lower = higher rank)
2. For each role in order, check for explicit grant or deny
3. First explicit decision found wins

This enables patterns like:
- `#announcements` rooms where `everyone` is denied `message.post` but `owner/admin/moderator` can still post (higher rank checked first), while everyone retains `message.post-in-thread` to discuss in threads
- Instance admin not being blocked by an `everyone` denial

**Testing implication:** Denying a permission on the `everyone` role does NOT block users with higher-rank roles (like `admin`). To test permission denial, deny on the user's actual highest-rank role or a role with equal/higher rank.

#### Permission Constant Naming

Permission constants follow the pattern `InstPerm{Category}{Action}` (singular nouns):

| Pattern | Example | Notes |
|---------|---------|-------|
| `InstPerm{Category}{Action}` | `InstPermSpaceCreate` | Singular category |
| `InstPermAdmin{Area}{Action}` | `InstPermAdminUsersView` | Admin permissions |
| `InstPermDM{Action}` | `InstPermDMWrite` | DM permissions |

**Common mistakes** (avoid these):
- `InstPermSpacesCreate` -> Use `InstPermSpaceCreate` (singular)
- `InstPermDMsWrite` -> Use `InstPermDMWrite` (no plural 's')
- `InstPermAdminAccessUsersView` -> Use `InstPermAdminUsersView`

The Go constants in `cli/internal/core/permissions.go` are the source of truth. Frontend TypeScript types are generated via `mise codegen-types`.

#### Permission String Naming

Permission strings use **hyphens** as word separators (e.g., `message.post-in-thread`, `message.edit-own`, `message.reply-in-thread`). Never use underscores in permission strings.

#### Built-in Permissions

| Permission | Description |
|------------|-------------|
| `space.manage` | Update space settings (name, description) |
| `space.delete` | Delete the space |
| `roles.manage` | Create/edit/delete roles |
| `roles.assign` | Assign roles to users |
| `members.invite` | Invite new members |
| `members.remove` | Remove members from space |
| `rooms.browse` | View list of rooms in space |
| `rooms.create` | Create new rooms |
| `rooms.manage` | Update/delete any room |
| `rooms.join` | Join existing rooms |

### GraphQL Authorization Reference

#### Queries

| Query | Auth Required | Additional Check |
|-------|---------------|------------------|
| `me` | No | Returns null if unauthenticated |
| `user(id)` | No | Public user profiles |
| `users` | Yes | Instance admin only |
| `spaces` | No | Discovery - lists all spaces |
| `space(id)` | No | Discovery - view any space |
| `room(spaceId, roomId)` | Yes | Room membership required |
| `roomEvents(...)` | Yes | Room membership required |
| `roomEvent(...)` | Yes | Room membership required |
| `admin` | Yes | Instance admin only |

#### Mutations

| Mutation | Auth Required | Additional Check |
|----------|---------------|------------------|
| `createUser` | No | Self-registration |
| `createSpace` | Yes | None (anyone can create) |
| `updateSpace` | Yes | `space.manage` |
| `joinSpace` | Yes | None |
| `leaveSpace` | Yes | None |
| `createRoom` | Yes | `rooms.create` |
| `joinRoom` | Yes | Space membership + `rooms.join` |
| `leaveRoom` | Yes | None |
| `postMessage` | Yes | Room membership + `message.post` (root) or `message.post-in-thread` (thread reply), + `message.reply` (if `inReplyTo` in room) or `message.reply-in-thread` (if `inReplyTo` in thread), + `message.echo` (if `alsoSendToChannel`) |
| `markRoomAsRead` | Yes | Room membership |
| `addReaction` | Yes | Room membership |
| `removeReaction` | Yes | Room membership |
| `deleteMessage` | Yes | Room membership + message ownership |
| `updateMyPresence` | Yes | None (sets caller's own presence) |

#### Subscriptions

| Subscription | Auth Required | Additional Check |
|--------------|---------------|------------------|
| `mySpaceEvents(spaceId)` | Yes | Space membership |
| `mySpaceLiveEvents(spaceId)` | Yes | Space membership |
| `myInstanceEvents` | Yes | None (user's own events) |
| `presenceUpdates(spaceId)` | Yes | Space membership |

#### Field Resolvers

| Field | Auth Required | Additional Check |
|-------|---------------|------------------|
| `Space.rooms` | Yes | Space membership + `rooms.browse` |
| `Space.memberCount` | No | Public count |
| `Space.roomCount` | No | Public count |
| `Space.assetCount` | No | Public count |
| `Room.members` | Yes | Room membership |
| `Room.hasUnread` | No | Returns false if unauthenticated |
| `User.spaces` | Yes | Self only (`caller.Id == obj.Id`) |
| `User.rooms` | Yes | Self only (`caller.Id == obj.Id`) |
| `User.avatarURL` | No | Public |
| `User.presenceStatus` | No | Public |

### Implementation Patterns

#### GraphQL Resolver with Permission Check
```go
func (r *mutationResolver) CreateRoom(ctx context.Context, input model.CreateRoomInput) (*Room, error) {
    user, err := requireAuth(ctx)
    if err != nil {
        return nil, err
    }

    // Check permission at GraphQL layer
    can, err := r.core.CanCreateRoom(ctx, user.Id, input.SpaceID)
    if err != nil {
        return nil, err
    }
    if !can {
        return nil, core.ErrPermissionDenied
    }

    // Core function assumes caller is authorized
    return r.core.CreateRoom(ctx, user.Id, input.SpaceID, input.Name, input.Desc)
}
```

#### Core Function (no authorization check)
```go
// CreateRoom creates a new room in a space.
// Authorization: Caller must verify CanCreateRoom before calling.
func (c *ChattoCore) CreateRoom(ctx context.Context, actorID, spaceID, name, desc string) (*Room, error) {
    // Business logic only - no permission check here
}
```

#### Authentication Helpers (in graph/authz.go)
```go
user, err := requireAuth(ctx)           // Returns authenticated user or error
user, err := requireSpaceMember(ctx, r.core, spaceID)  // + space membership
user, err := requireRoomMember(ctx, r.core, spaceID, roomID)  // + room membership
```

#### Self-Only Access Check
```go
if caller.Id != obj.Id {
    return nil, fmt.Errorf("access denied: cannot view other users' data")
}
```

### Customizable Permissions

Default member permissions (`rooms.browse`, `rooms.create`, `rooms.join`) can be revoked from the member role. When implementing or modifying permission checks:

1. **Always use the RBAC engine** - Never hardcode permission grants based on role names or "default" lists
2. **Test both grant and revoke** - Permissions must work when granted AND when revoked
3. **Follow the instance RBAC pattern** - Use `engine.RoleHasPermission(ctx, RoleMember, permStr)` to check actual KV state

**Anti-pattern (avoid):**
```go
// BAD: Hardcoded bypass that ignores actual role permissions
if isMember && isDefaultPermission(perm) {
    return true, nil  // Bypasses RBAC engine!
}
```

**Correct pattern:**
```go
// GOOD: Always check actual role permissions via RBAC engine
if isMember {
    hasPerm, err := engine.RoleHasPermission(ctx, RoleMember, string(perm))
    if hasPerm {
        return true, nil
    }
}
```

### Instance Admin vs Space Admin

Two separate authorization concepts:

- **Instance admin**: Configured via `admin.emails` in `chatto.toml`. Can access `/admin` routes to view system-wide data.
- **Space admin**: Per-space role (`RoleAdmin` in permissions.go). Can manage a specific space's settings, rooms, and members.

These are independent - a space admin is not automatically an instance admin and vice versa.

Instance admins are configured via `admin.emails` in `chatto.toml`. They have access to:

- `/admin` routes in the frontend
- `Query.admin` and `Query.users` in GraphQL
- System monitoring data (NATS stats, streams, KV buckets)

### Privacy Boundary

Instance admins can see operational metadata but NOT user content:

| Can See                            | Cannot See       |
| ---------------------------------- | ---------------- |
| User list (login, email, avatar)   | Message content  |
| Space/room names and member counts | Private messages |
| NATS/JetStream metrics             | File contents    |
| System configuration               | User passwords   |

This boundary is intentional. If message visibility is needed for moderation, it should be a separate, auditable feature with explicit consent.

### Admin Backend Authorization

Admin queries use a nested `admin` type pattern. The `Query.admin` resolver checks authorization once and returns `nil` for non-admins:

```go
func (r *queryResolver) Admin(ctx context.Context) (*model.AdminQueries, error) {
    user := auth.ForContext(ctx)
    if user == nil {
        return nil, nil // Not authenticated
    }
    if !isConfigAdmin(ctx, r.core, r.adminConfig, user.Id) {
        return nil, nil // Not an admin
    }
    // Return populated AdminQueries...
}
```

The `isConfigAdmin` helper checks if any of the user's _verified_ emails match the `admin.emails` list. Unverified/pending emails are never matched.

All fields under `admin` (users, spaces, systemInfo) don't need individual auth checks - the parent resolver handles it.

### Admin Configuration

```toml
[admin]
emails = ["admin@example.com", "ops@example.com"]
```

Users are granted instance admin access if any of their verified email addresses matches an entry in this list. The `isConfigAdmin` helper performs the matching - only verified emails are considered, never pending/unverified ones.

---

## Backup & Restore

### Architecture

- `chatto backup` connects to a running NATS server via client config, snapshots all JetStream streams using `jsm.go`'s `SnapshotToDirectory`, and creates a `.tar.gz` archive with a `manifest.json`. Use `--encrypt` for age-encrypted archives (`.tar.gz.age`).
- `chatto restore` extracts the archive and restores each stream using `jsm.go`'s `RestoreSnapshotFromDirectory`. Auto-detects age-encrypted archives. For embedded NATS setups, it starts a temporary in-process NATS server. For external NATS, it connects via the client config.

### Key Files

- `cli/cmd/backup.go` — Backup command, tar/gzip utilities, encrypted archive support, skip list, manifest types
- `cli/cmd/restore.go` — Restore command with conflict handling, age detection, and embedded NATS support
- `cli/cmd/keys.go` — Encryption key export/import with age encryption

### Encryption

All encryption (backup archives and key exports) uses `filippo.io/age` with passphrase-based scrypt recipients. This is the same format as the `age` CLI tool — files are interoperable.

Key functions:
- `createEncryptedTarGz()` / `extractEncryptedTarGz()` — streaming backup encryption
- `encryptKeysToFile()` / `decryptKeysFromFile()` — key export encryption
- `isAgeEncrypted()` — detects age header for auto-detection in restore
- `getPassphrase(flagValue, prompt, confirm)` — shared passphrase input (flag or interactive)

The tar functions are split into streaming versions (`writeTarGz`/`readTarGz` accepting `io.Writer`/`io.Reader`) and file wrappers (`createTarGz`/`extractTarGz`). This enables chaining: file -> age -> gzip -> tar.

### Stream Skip List

The `skipReason()` function in `backup.go` determines which streams are excluded. When adding new KV buckets or streams to core, consider whether they should be backed up:

| Should backup | Should skip |
|---------------|-------------|
| User data, messages, config | Caches (regeneratable) |
| Roles, permissions, memberships | Ephemeral/memory streams |
| Assets (avatars, attachments) | Security-sensitive (encryption keys, auth tokens) |

If you add a new memory-only or cache bucket, add it to `skipReason()`.

### Encryption Keys

Encryption keys (`KV_ENCRYPTION_KEYS`) are intentionally excluded from data backups. This is a security design choice — backups contain only encrypted data, not the keys to decrypt it.

Separate key management commands exist:
- `chatto keys export -o keys.backup` — Exports all per-user encryption keys, encrypted with age
- `chatto keys import keys.backup` — Imports keys; skips users that already have a key (safe to re-run)

Key files: `cli/cmd/keys.go`, `cli/cmd/keys_test.go`

The export file format is version 2: an age-encrypted JSON payload containing a `KeyExport` struct with version, timestamp, and key array.

### Manifest Format (v1)

```json
{
  "version": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "streams": [
    {"name": "KV_INSTANCE", "type": "kv", "messages": 42, "bytes": 1024},
    {"name": "KV_USER_PRESENCE", "type": "skipped", "messages": 0, "bytes": 0}
  ],
  "stats": {
    "total_streams": 10,
    "total_bytes": 102400,
    "duration_ms": 500,
    "skipped": 3,
    "failed": 0
  }
}
```

### Restore Conflict Modes

- `--conflict=error` (default): Fail if any stream exists. Safe for fresh restores.
- `--conflict=skip`: Skip existing streams. Useful for partial restore.
- `--conflict=overwrite`: Delete and recreate. Destructive but complete.

---

## Documentation Website

> Applies to files in `docs-website/`.

The docs website lives in `docs-website/` and is built with [Starlight](https://starlight.astro.build/) (Astro).

### Keeping Docs in Sync

The docs website must stay in sync with the codebase. When adding, changing, or removing user-facing features, configuration options, or deployment behavior, update the corresponding docs pages:

- New environment variables or TOML options -> add to `docs-website/src/content/docs/reference/environment-variables.mdx`
- New or changed features -> update relevant guide in `docs-website/src/content/docs/guides/`
- Changed config defaults or semantics -> update both the reference and any guides that mention the option

If you notice documentation that is out of date or inconsistent with the code, alert the user about the drift before proceeding.

### Assumptions

- The code repository is public and all binaries and Docker images (including `ghcr.io/hmans/chatto`) are publicly available. Don't include setup steps for repository access or container registry authentication.
- Refer to calls as **"voice and video calls"** or just **"calls"** — never "voice calls" alone. LiveKit handles both audio and video.

### Starlight Components

Use built-in Starlight components where appropriate. All are imported from `@astrojs/starlight/components`:

| Component | Use for |
|-----------|---------|
| `Steps` | Numbered setup/tutorial sequences |
| `Aside` | Callouts — `tip`, `note`, `caution`, `danger` |
| `FileTree` | Showing directory/file structures |
| `LinkCard` | Cross-references to other docs pages |
| `CardGrid` | Laying out multiple `LinkCard`s side-by-side |
| `Tabs` / `TabItem` | Showing alternatives (e.g., dev vs. prod config) |

### Sidebar Configuration

The sidebar is configured in `docs-website/astro.config.mjs` under `starlight.sidebar`. When adding new pages, add them to the appropriate section there.

### Avoiding Duplication

Prefer linking to dedicated guide pages rather than repeating detailed instructions in multiple places. For example, the Docker Compose page should link to the S3 storage guide rather than documenting S3 configuration inline. Use `LinkCard` components for these cross-references.

### Writing Style

- **Direct and concise.** Lead with what the reader needs to know, not background. Skip "In this guide, we will..." preambles.
- **Second person, present tense.** "You can run multiple replicas" not "One can run" or "The user runs."
- **Confident tone.** State facts plainly. Avoid hedging ("might", "perhaps", "it should be noted that").
- **Short paragraphs.** One idea per paragraph. Use tables and lists over long prose.
- **Show, then explain.** Put the config example first, then explain what it does — not the other way around for simple options.

### Terminology

- **"instance"** refers to a Chatto deployment (the logical entity with users, spaces, data). Don't use "instance" to mean a running process or replica.
- **"process"** or **"replica"** for individual running copies of the Chatto binary.
- **"calls"** or **"voice and video calls"** — never "voice calls" alone.
- Don't recommend MinIO — it's dead. Use Cloudflare R2, Wasabi, Backblaze B2, or AWS S3 as example providers.

### Content Conventions

- Use `example.com` as the placeholder domain (e.g., `chat.example.com`, `livekit.chat.example.com`)
- Use `<generate-me>` as a placeholder for secrets that need to be generated
- Show both TOML config and environment variable alternatives where applicable (use `Tabs` component)
- Link to the environment variables reference for full option lists rather than duplicating them

### Font Sizing

Don't shrink text. Use the base font size for all readable content — titles, descriptions, body text. Only use smaller sizes (`text-xs`, `0.6rem`, etc.) for badges, labels, port numbers, and other metadata that isn't meant to be read as prose.

### Architecture Diagrams

SVG architecture diagrams live in `docs-website/src/assets/` and are imported as raw strings (`?raw`) for inline rendering — this is required for SVG animations to work (an `<img>` tag won't animate).

#### Design Patterns

- **Dark/light mode**: Use `@media (prefers-color-scheme: light)` inside the SVG `<style>` to provide both color schemes. Dark mode is the default.
- **Box colors**: Each service type has its own subtle fill + border color (e.g., `.box-chatto`, `.box-nats`). Keep these muted — they shouldn't compete with the animated dots.
- **Connection lines**: Use `.conn` (solid) for persistent connections and `.conn-dash` (dashed) for direct/UDP connections that bypass the proxy.

#### Animation Guidelines

- Use `<animateMotion>` with `<mpath>` to move dots along connection paths.
- **Easing**: Always use `calcMode="spline"` with `keySplines="0.4 0 0.2 1"` for a smooth ease-in-out feel. Never use linear (`calcMode="linear"` or default).
- **Bidirectional connections** (e.g., Chatto<->NATS): Use a single dot that bounces back and forth with `keyPoints="0;1;0"` and `keyTimes="0;0.5;1"` (two spline segments). Don't use two separate dots on parallel offset paths — it looks janky.
- **Unidirectional connections** (e.g., Browser->Caddy): Use one or two dots with staggered `begin` offsets traveling in the same direction.
- **Dot sizing**: Use `r="3"` for primary traffic dots, `r="2.5"` for secondary/API dots. Use `opacity="0.7"` to visually de-emphasize less important connections.
- **Dot colors**: `.dot` (sky blue) for main HTTP/WS traffic, `.dot-yellow` for media/WebRTC traffic, `.dot-blue` for internal messaging (NATS).
