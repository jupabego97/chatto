---
# chatto-5g7z
title: Migrate NATS subjects from per-message to per-room/per-thread cardinality
status: draft
type: feature
priority: high
created_at: 2026-03-17T07:51:58Z
updated_at: 2026-03-17T07:52:20Z
---

Remove event IDs from NATS JetStream subjects to reduce subject index cardinality from O(messages) to O(rooms + threads).

## Motivation

Every message creates a unique NATS subject (`msg.{eventId}`), so the JetStream subject index grows linearly with total message count. This is the most expensive kind of cardinality for NATS memory usage.

## New Subject Patterns

| Type | Current | New |
|------|---------|-----|
| Root message | `space.{s}.room.{r}.msg.{eventId}` | `space.{s}.room.{r}.msg.root` |
| Thread reply | `space.{s}.room.{r}.msg.{rootEventId}.replies.{eventId}` | `space.{s}.room.{r}.msg.thread.{rootSequenceId}` |
| Meta | unchanged | unchanged |

Cardinality drops from **one subject per message** to **one per room + one per thread**.

Filter patterns: `msg.>` (all), `msg.root` (roots), `msg.thread.>` (all threads), `msg.thread.{rootSeq}` (specific thread).

## Key Design Decision: Sequence IDs Replace Event IDs

Thread subjects use the root message's JetStream sequence ID (uint64) instead of its event ID (NanoID string). This means:

- `inThread` and `inReplyTo` become `uint64` in proto and numeric `ID` strings in GraphQL
- All thread metadata, follows, and last-opened KV keys switch from event ID to sequence ID
- `GetRoomEventByEventID` is deleted — all lookups use `GetRoomEvent(sequenceID uint64)`
- The `roomEventByEventId` GraphQL query is removed
- OCC becomes functional (detects concurrent posts to same room/thread) instead of being a no-op

JetStream sequence IDs are immutable after creation under normal operation.

## Scope

### Backend (breaking changes to proto, GraphQL schema, NATS subjects)

- [ ] `subjects.go` — new subject constructors, remove event ID params, delete `SpaceRoomThreadLookup`/`ParseEventIDFromSubject`
- [ ] Proto — `in_thread`/`in_reply_to` from `string` to `uint64`, rename echo fields on `MessagePostedEvent`, notification protos, `ThreadMetadata.root_sequence_id`
- [ ] GraphQL schema — remove `roomEventByEventId`, rename `threadRootEventId` to `threadRootSequenceId` everywhere, notification field renames
- [ ] `rooms.go` — `PostMessage(inThread, inReplyTo uint64)`, delete `GetRoomEventByEventID`, switch all thread functions to sequence-based keys
- [ ] Resolvers — parse sequence IDs from GraphQL `ID` strings, populate `SequenceId` on inner `MessagePostedEvent` for thread metadata resolvers
- [ ] Push notifications — switch from event ID to sequence ID in notification payloads and URLs
- [ ] Tests — update all PostMessage calls, thread assertions, notification field references

### Frontend (~15 files)

- [ ] GraphQL operations — `roomEventsAround(sequenceId:)`, `threadEvents(threadRootSequenceId:)`, typing indicator mutation
- [ ] GraphQL fragments — `MessagePostedEvent` echo fields, `UserTypingEvent`, `ThreadFollowChangedEvent`, notification types
- [ ] `MessageEvent.svelte` — replace deleted `roomEventByEventId` query for reply previews (use `roomEvent(sequenceId:)` or resolve from loaded events)
- [ ] `MessageComposer.svelte` — pass `sequenceId` instead of event ID for `inThread`/`inReplyTo`
- [ ] State/event buses — `replyState`, `notifications`, `instanceEventBus`, `spaceEventBus` field renames
- [ ] Navigation — thread URLs, notification click URLs, `?highlight=` params all use sequence IDs
- [ ] `threads/+page.svelte` — followed threads list field rename
- [ ] Run `mise codegen-frontend` to regenerate TypeScript types

### Docs

- [ ] Update `docs/ARCHITECTURE.md` subject tables and filtering examples
- [ ] Update `.claude/rules/nats-subjects.md`
- [ ] ADR documenting the migration rationale

## Notes from Prototype

A full backend prototype was built and validated (all Go tests passing). Key learnings:

- The `MessagePostedEvent` proto needs a runtime-only `sequence_id` field (like `event_id`) so nested resolvers can look up thread metadata by sequence ID
- `MessageUpdatedEvent` in `live_event.proto` also has `in_thread`/`in_reply_to` fields that need the same type change
- The `NotificationCreatedEvent` fields need `@goField(forceResolver: true)` in the GraphQL schema since gqlgen can't auto-bind `ID` (string) to `uint64`
- Thread metadata, follow, and last-opened KV keys format the uint64 sequence ID as a string for the key
- The frontend `MessageEvent.svelte` reply preview is the trickiest part — it uses the now-deleted `roomEventByEventId` query
