---
# chatto-v9rm
title: Room Muting
status: draft
type: feature
created_at: 2026-01-20T11:50:19Z
updated_at: 2026-01-20T11:50:19Z
parent: chatto-bq1a
---

Allow users to mute specific rooms to suppress all notifications from those rooms.

## Overview

Sometimes users want to be a member of a room but don't want to receive notifications from it. Room Muting provides a simple on/off toggle to silence all notifications from a specific room while remaining a member.

## Use Cases

- High-traffic channels (#random, #watercooler) that aren't urgent
- Rooms for teams the user monitors but isn't actively involved in
- Temporary mute during focused work sessions
- Archive-style rooms for reference but not active discussion

## User Experience

### Muting a Room

1. **Room header menu**: Click room name → "Mute room" option
2. **Room list context menu**: Right-click room → "Mute room"
3. **Room settings**: Toggle in room settings panel (if we have one)

### Visual Indicators

- Muted rooms show a mute icon (🔇) next to the name in room list
- Muted rooms are visually dimmed/grayed out slightly
- Muted rooms DON'T show unread indicators (since user opted out)
- Tooltip on mute icon: "This room is muted"

### Unmuting

- Same locations as muting: "Unmute room" replaces "Mute room"
- Unmuting immediately restores normal notification behavior
- Previous unread messages don't retroactively trigger notifications

## What Gets Muted

When a room is muted, the user will NOT receive:
- @mention notifications from that room
- Group mention notifications (@everyone, @here)
- Thread reply notifications for threads in that room
- Unread indicators for that room
- Any future notification types added to that room

The user WILL still:
- See messages when they open the room
- Be able to post messages
- Appear in room member list
- Receive DMs (DMs are separate, not room-based)

## Technical Design

### Storage

Per-user mute status in `SPACE_{spaceId}_RUNTIME` KV bucket:

```
Key: room_mute.{userId}.{roomId}
Value: RoomMuteEntry proto (or simple bool/timestamp)
```

Alternatively, store in user's space membership:
- Pro: Single source of truth for membership + preferences
- Con: More complex membership proto, harder to query muted rooms

Recommendation: Separate key for simplicity and query efficiency.

### Proto Definition

```protobuf
message RoomMuteEntry {
  string user_id = 1;
  string room_id = 2;
  google.protobuf.Timestamp muted_at = 3;  // When muted, useful for analytics
}
```

Or simpler: just presence of the key means muted (value can be empty or timestamp).

### Core Functions

```go
// MuteRoom mutes a room for a user
func (c *ChattoCore) MuteRoom(ctx context.Context, userID, spaceID, roomID string) error

// UnmuteRoom unmutes a room for a user
func (c *ChattoCore) UnmuteRoom(ctx context.Context, userID, spaceID, roomID string) error

// IsRoomMuted checks if a room is muted for a user
func (c *ChattoCore) IsRoomMuted(ctx context.Context, userID, spaceID, roomID string) (bool, error)

// GetMutedRooms returns all muted rooms for a user in a space
func (c *ChattoCore) GetMutedRooms(ctx context.Context, userID, spaceID string) ([]string, error)
```

### GraphQL API

```graphql
type Room {
  # Existing fields...
  isMuted: Boolean! @goField(forceResolver: true)  # Viewer's mute status
}

type Mutation {
  muteRoom(spaceId: ID!, roomId: ID!): Room!
  unmuteRoom(spaceId: ID!, roomId: ID!): Room!
}

# Alternatively, a toggle:
type Mutation {
  setRoomMuted(spaceId: ID!, roomId: ID!, muted: Boolean!): Room!
}
```

### Integration with Notification System

All notification paths check mute status first:

```go
func (c *ChattoCore) shouldNotifyUser(...) (bool, error) {
    // First check: Is the room muted?
    if muted, _ := c.IsRoomMuted(ctx, userID, spaceID, roomID); muted {
        return false, nil  // Skip all notifications for muted rooms
    }
    // ... rest of notification preference logic
}
```

### Integration with Unread Indicators

When checking `hasUnread` for a room:

```go
func (r *roomResolver) HasUnread(ctx context.Context, obj *corev1.Room) (bool, error) {
    user := auth.ForContext(ctx)
    if user == nil {
        return false, nil
    }
    
    // Don't show unread for muted rooms
    if muted, _ := r.core.IsRoomMuted(ctx, user.Id, obj.SpaceId, obj.Id); muted {
        return false, nil
    }
    
    return r.core.HasUnread(ctx, obj.SpaceId, obj.Id, user.Id)
}
```

Same for `hasMention`:

```go
func (r *roomResolver) HasMention(ctx context.Context, obj *corev1.Room) (bool, error) {
    // ... auth check
    if muted, _ := r.core.IsRoomMuted(ctx, user.Id, obj.SpaceId, obj.Id); muted {
        return false, nil
    }
    return r.core.HasMention(ctx, obj.SpaceId, obj.Id, user.Id)
}
```

## UI Implementation

### Room List Changes

```svelte
<!-- In RoomList.svelte -->
{#each rooms as room}
  <a href={...} class={[
    'sidebar-item',
    room.isMuted ? 'opacity-60' : '',
    // ... other classes
  ]}>
    <span class="sidebar-icon">#</span>
    <span class="flex-1 truncate">{room.name}</span>
    
    {#if room.isMuted}
      <span class="icon-[uil--volume-mute] text-muted" title="Muted"></span>
    {:else if room.hasMention && room.id !== activeRoomId}
      <span class="h-2 w-2 rounded-full bg-warning"></span>
    {:else if room.hasUnread && room.id !== activeRoomId}
      <span class="h-2 w-2 rounded-full bg-primary"></span>
    {/if}
  </a>
{/each}
```

### Room Header Menu

Add mute/unmute option to room actions dropdown:

```svelte
<DropdownMenu>
  <!-- ... other items -->
  {#if room.isMuted}
    <DropdownItem onclick={unmute}>
      <span class="icon-[uil--volume]"></span>
      Unmute room
    </DropdownItem>
  {:else}
    <DropdownItem onclick={mute}>
      <span class="icon-[uil--volume-mute]"></span>
      Mute room
    </DropdownItem>
  {/if}
</DropdownMenu>
```

### Query Update

Update room queries to include mute status:

```graphql
query GetMyRoomsInSpace($spaceId: ID!) {
  me {
    roomMemberships(spaceId: $spaceId) {
      room {
        id
        name
        hasUnread
        hasMention
        isMuted  # Add this
      }
    }
  }
}
```

## Implementation Tasks

- [ ] Add RoomMuteEntry proto (or decide on simple key presence)
- [ ] Implement MuteRoom / UnmuteRoom / IsRoomMuted in Core
- [ ] Add `isMuted` field to Room GraphQL type
- [ ] Add muteRoom / unmuteRoom mutations
- [ ] Update hasUnread resolver to check mute status
- [ ] Update hasMention resolver to check mute status
- [ ] Update notification paths to check mute status
- [ ] Add mute icon to room list for muted rooms
- [ ] Add visual dimming for muted rooms
- [ ] Add mute/unmute to room header menu
- [ ] Update room query to fetch isMuted
- [ ] Add E2E test: muting prevents notifications
- [ ] Add E2E test: muted rooms don't show unread indicator
- [ ] Add E2E test: mute/unmute toggle works

## Edge Cases

- **Mute then leave room**: Mute entry can be orphaned (harmless, clean up on rejoin or periodic sweep)
- **Mute then rejoin**: Should mute status persist? Probably yes (user preference)
- **DM rooms**: Can DM rooms be muted? Yes, treat like any other room
- **Default room**: Can the default room be muted? Yes, but maybe warn user

## Design Decisions

- **Mute persistence**: Mute status persists until explicitly unmuted
- **No "mute until" option**: Keep it simple for V1; timed mute can be added later
- **No "mute categories"**: Full mute only; partial mute is handled by Notification Preferences

## Dependencies

- Notification system (for integration)
- Room list UI (for mute indicator)