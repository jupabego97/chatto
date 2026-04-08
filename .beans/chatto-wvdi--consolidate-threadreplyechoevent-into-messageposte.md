---
# chatto-wvdi
title: Consolidate ThreadReplyEchoEvent into MessagePostedEvent
status: in-progress
type: task
priority: high
created_at: 2026-04-05T18:43:03Z
updated_at: 2026-04-05T19:50:12Z
---

Fold ThreadReplyEchoEvent into MessagePostedEvent at all layers (proto, GraphQL, Go, frontend). One message type everywhere.

## Context

ThreadReplyEchoEvent is a separate proto message and GraphQL type for "also send to channel" thread replies. It duplicates most of MessagePostedEvent's fields. Since we're still in alpha, we can make breaking changes to persisted data (old echo events in JetStream will become unreadable).

## Plan

### Proto changes
- [ ] Add `echo_of_event_id` (string, optional) to `MessagePostedEvent` proto
- [ ] Add `thread_root_event_id` (string, optional) to `MessagePostedEvent` proto
- [ ] Remove `ThreadReplyEchoEvent` message from `space_event.proto`
- [ ] Remove `thread_reply_echo = 403` from `SpaceEvent` oneof, reserve the field number
- [ ] Run `mise codegen-cli` to regenerate proto Go code

### Core (Go) changes
- [ ] Update echo publish code to create `MessagePostedEvent` with echo fields instead of `ThreadReplyEchoEvent`
- [ ] Remove `unwrapSpaceEvent` case for `ThreadReplyEchoEvent`
- [ ] Remove `IsSpaceEventType()` marker on `ThreadReplyEchoEvent` in `graphql.go`
- [ ] Update subscription handler room ID extraction (remove `ThreadReplyEchoEvent` case)
- [ ] Update `resolveMessageBodyKey` to remove `ThreadReplyEchoEvent` case (echoes are now `MessagePostedEvent`)
- [ ] Update event helpers, link preview resolvers, etc.

### GraphQL schema changes
- [ ] Remove `ThreadReplyEchoEvent` type from `events.graphqls`
- [ ] Remove from `SpaceEventType` union
- [ ] Add `echoOfEventId: ID` and `threadRootEventId: ID` to `MessagePostedEvent`
- [ ] Strip dead content fields from `MessageUpdatedEvent` (body, attachments, reactions, inReplyTo — unused by frontend)
- [ ] Run `mise codegen-cli`

### Frontend changes
- [ ] Update RoomEvent.svelte fragment — remove `ThreadReplyEchoEvent` fragment, add echo fields to `MessagePostedEvent`
- [ ] Replace all `__typename === 'ThreadReplyEchoEvent'` checks with `echoOfEventId != null`
- [ ] Update RoomEventsPane.svelte live event handler (remove separate echo handler, merge into MessagePostedEvent handler)
- [ ] Update ThreadPane.svelte
- [ ] Update MessageEvent.svelte — one rendering path, echo is just a "flavor"
- [ ] Update event correlation logic (echoOfEventId matching)
- [ ] Update messageGrouping.spec.ts
- [ ] Run `mise codegen-frontend`

### Optional follow-up
- Consider extracting `MessageContent` type for lazy-loaded fields (body, attachments, reactions, inReplyTo, updatedAt) — only if the flat field count on MessagePostedEvent feels unwieldy

## Design decisions
- Echo fields (`echoOfEventId`, `threadRootEventId`) stay flat on the event — they're event metadata, not message content
- Old ThreadReplyEchoEvent data in JetStream becomes unreadable (alpha-stage, acceptable)
- `MessageUpdatedEvent` becomes a pure notification event (no content fields)
