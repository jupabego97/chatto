# Unified DM Inbox

## The Basics

- DM conversations from all connected instances are displayed in a single inbox sidebar, rather than being scoped to a single instance.
- The DM route lives at `/chat/dm/` (outside the `[instanceId]` route tree) so the conversation list sidebar stays mounted across instance switches without remounting.
- Each conversation entry shows which instance it belongs to (hostname label), visible when multiple instances are connected.
- Conversations load progressively from all instances in parallel; a slow or unreachable instance doesn't block the others.

## Routing

- DM inbox landing page: `/chat/dm`
- DM conversation: `/chat/dm/{instanceSegment}/{conversationId}`
  - `instanceSegment` is `-` for the home instance, or the remote hostname (e.g., `chat.example.com`)
- The `[instanceSegment]/+layout.svelte` provides the instance-specific connection context (`provideConnection`), current user (`setCurrentUser`), and wraps children in a `SpaceEventProvider` for the DM space.
- The `/chat/dm` landing page shows a placeholder prompting the user to select a conversation.

## DMConversationList

- Lives in `dm/+layout.svelte` (above the `[instanceSegment]` layout) so it stays mounted regardless of which instance's conversation is open.
- Queries `GetDMConversationsForList` on every registered instance via `graphqlClientManager.getClient(instance.id)`.
- Merges results into a single `$state` array, tagged with `instanceId` and `instanceHostname`.
- Subscribes to DM notifications from all instance event buses (`instanceEventBusManager`) for `NewDirectMessageNotificationEvent` to bump conversations when new DMs arrive.
- Subscribes to DM space events (`SpaceEventBusSubscription`) per instance to catch all messages (including your own) for immediate sidebar updates.
- Shows unread indicator (warning dot) per conversation; clears when the conversation is opened.
- Instance hostname labels only render when `instanceRegistry.instances.length > 1`.

## Real-Time Updates

- **SpaceEventProvider** in `[instanceSegment]/+layout.svelte` subscribes to `mySpaceEvents(spaceId: "DM")` on the active conversation's instance. This delivers message events to the Room's `RoomEventsPane` via the `SpaceEventBus` context.
- **DMConversationList** maintains separate per-instance subscriptions for sidebar updates (conversation ordering, unread state).
- **Instance event bus** handlers catch `NewDirectMessageNotificationEvent` for cross-instance DM notifications.
- The SpaceEventProvider tracks `reconnectCount` to explicitly restart the subscription after WebSocket reconnections.

## Starting a DM

- DMs are started from user context menus in instance-specific room views (member list clicks, @mention clicks, message author clicks).
- `startDMWith()` in `$lib/dm/startDM.ts` uses the correct instance's GraphQL client based on the `instanceId` parameter (not hardcoded to `homeClient`).
- There is no "new DM" button in the unified inbox itself; starting a DM requires navigating through an instance-scoped space view.

## Authorization

- The DM space uses permission-based access (`dm.view`, `dm.write`) rather than space membership. The backend's `requireSpaceMember` has a special case for `IsDMSpace(spaceID)` that checks `HasInstancePermission` instead of `SpaceMembershipExists`.
- The DM inbox is gated by the home instance's `canViewDMs` permission in `dm/+layout.svelte`.
- Individual DM rooms use standard room membership checks.

## Key Files

| File | Purpose |
|------|---------|
| `frontend/src/routes/chat/dm/+layout.svelte` | DM inbox layout with permission guard and sidebar |
| `frontend/src/routes/chat/dm/+page.svelte` | Landing page (placeholder) |
| `frontend/src/routes/chat/dm/[instanceSegment]/+layout.svelte` | Instance context, connection, SpaceEventProvider |
| `frontend/src/routes/chat/dm/[instanceSegment]/[conversationId]/+page.svelte` | Renders Room component |
| `frontend/src/lib/dm/DMConversationList.svelte` | Unified conversation list with multi-instance queries and subscriptions |
| `frontend/src/lib/dm/startDM.ts` | Instance-aware DM creation |
| `cli/internal/graph/authz.go` | `requireSpaceMember` DM special case |
| `cli/internal/core/dm.go` | DM space constants, `IsDMSpace`, `FindOrCreateDM` |
