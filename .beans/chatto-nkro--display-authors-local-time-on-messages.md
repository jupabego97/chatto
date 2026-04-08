---
# chatto-nkro
title: Display author's local time on messages
status: draft
type: feature
priority: normal
created_at: 2026-01-24T23:06:52Z
updated_at: 2026-01-24T23:07:03Z
---

When a user sends a message, capture their timezone offset and display it to recipients so they can see what time it was for the author when they sent it. Useful for async communication across timezones.


## Design Decisions

**Store UTC offset in minutes (not IANA timezone name)**
- Simpler storage (int32 vs string)
- The offset at message creation time is what matters
- DST changes are irrelevant: if someone sent at 10:00 AM their time, we show 10:00 AM
- JavaScript's `getTimezoneOffset()` returns minutes, directly usable

**Backward compatibility**
- Default to `0` for old messages
- Frontend detects `0` and doesn't show author's time for legacy messages

## Implementation

### Backend

- [ ] Add `author_timezone_offset_minutes` field to `Event` message in `proto/chatto/core/v1/event.proto`
- [ ] Add `timezone_offset_minutes` to `PostMessageRequest` in `proto/chatto/core_api/v1/messages.proto`
- [ ] Add `authorTimezoneOffsetMinutes: Int!` to Event type in `cli/internal/graph/events.graphqls`
- [ ] Add `timezoneOffsetMinutes: Int` to PostMessageInput in `cli/internal/graph/mutation.graphqls`
- [ ] Run `mise codegen-cli`
- [ ] Update `PostMessage` in `cli/internal/core/rooms.go` to accept and store timezone offset
- [ ] Pass timezone offset through NATS API service (`cli/internal/core_api/messages_service.go`)
- [ ] Pass timezone offset in GraphQL resolver (`cli/internal/graph/mutation.resolvers.go`)

### Frontend

- [ ] Run `mise codegen-frontend`
- [ ] Send `timezoneOffsetMinutes: new Date().getTimezoneOffset()` in ChatInput.svelte
- [ ] Add `authorTimezoneOffsetMinutes` to RoomEventView fragment in RoomEvent.svelte
- [ ] Display author's local time in MessageEvent.svelte when different from viewer's time

### Testing

- [ ] Add unit tests for timezone offset handling in PostMessage
- [ ] Run full test suite (`mise test`)
- [ ] Manual test: verify author's local time displays correctly

## Display Logic

Show author's time only when:
1. `authorTimezoneOffsetMinutes !== 0` (not legacy/UTC)
2. Author's formatted time differs from viewer's formatted time

Format: `14:30 (09:30 their time)`

## Helper Function

```typescript
function formatAuthorLocalTime(createdAt: string, offsetMinutes: number): string {
  const utcDate = new Date(createdAt);
  const authorLocal = new Date(utcDate.getTime() - offsetMinutes * 60 * 1000);
  return authorLocal.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    timeZone: 'UTC'
  });
}
```