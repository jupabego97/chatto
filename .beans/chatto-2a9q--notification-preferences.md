---
# chatto-2a9q
title: Notification Preferences
status: draft
type: feature
created_at: 2026-01-20T11:49:37Z
updated_at: 2026-01-20T11:49:37Z
parent: chatto-bq1a
---

User settings to control what notifications they receive and how they receive them.

## Overview

As the notification system grows, users need fine-grained control over their notification experience. This feature provides a settings page where users can configure their notification preferences at global and per-space levels.

## Preference Levels

### Global Preferences (User Settings)
- Apply across all spaces unless overridden
- Set in `/chat/settings/notifications`

### Per-Space Preferences
- Override global settings for a specific space
- Set in space settings or via space menu

### Per-Room Preferences
- Covered by "Room Muting" feature (separate bean)
- Focus here is space-level and global

## Notification Categories

### @mentions
- Someone mentions you with @username
- Default: ON (all spaces)
- Options: All / None

### Group Mentions
- @everyone, @here, @role mentions
- Default: ON (all spaces)
- Options: All / @everyone only / None

### Direct Messages
- New DM messages
- Default: ON
- Options: All / None
- Note: Can't be configured per-space (DMs are global)

### Thread Replies
- Replies to threads you've participated in
- Default: ON (followed threads only)
- Options: All / Followed only / None

### All Messages (Per-Space)
- Every message in spaces you're in
- Default: OFF (would be overwhelming)
- Options: ON / OFF
- Warning shown if enabled

## UI Design

### Settings Page Structure

```
/chat/settings/notifications

┌─────────────────────────────────────────────────┐
│ Notification Preferences                         │
├─────────────────────────────────────────────────┤
│                                                  │
│ ## Global Settings                               │
│                                                  │
│ @mentions                    [All ▾]             │
│ Someone mentions you by name                     │
│                                                  │
│ Group mentions               [All ▾]             │
│ @everyone, @here, @role pings                    │
│                                                  │
│ Direct messages              [All ▾]             │
│ New messages in DMs                              │
│                                                  │
│ Thread replies               [Followed only ▾]   │
│ Replies in threads you've joined                 │
│                                                  │
│ ─────────────────────────────────────            │
│                                                  │
│ ## Per-Space Overrides                           │
│                                                  │
│ ┌─────────────────────────────────────────┐      │
│ │ Space Name          [Use Global ▾]      │      │
│ │ Another Space       [All mentions ▾]    │      │
│ │ + Add space override                    │      │
│ └─────────────────────────────────────────┘      │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Quick Access
- "Notification Settings" link in Notification Center panel
- "Mute space" option in space dropdown menu (sets to "None")

## Technical Design

### Storage

User preferences in `INSTANCE` KV bucket:

```
Key: user_notification_prefs.{userId}
Value: NotificationPreferences proto

Key: user_space_notification_prefs.{userId}.{spaceId}
Value: SpaceNotificationPreferences proto
```

### Proto Definitions

```protobuf
message NotificationPreferences {
  MentionPreference mentions = 1;
  GroupMentionPreference group_mentions = 2;
  DMPreference dms = 3;
  ThreadPreference threads = 4;
}

enum MentionPreference {
  MENTION_PREF_ALL = 0;
  MENTION_PREF_NONE = 1;
}

enum GroupMentionPreference {
  GROUP_MENTION_PREF_ALL = 0;
  GROUP_MENTION_PREF_EVERYONE_ONLY = 1;
  GROUP_MENTION_PREF_NONE = 2;
}

enum DMPreference {
  DM_PREF_ALL = 0;
  DM_PREF_NONE = 1;
}

enum ThreadPreference {
  THREAD_PREF_ALL = 0;
  THREAD_PREF_FOLLOWED_ONLY = 1;
  THREAD_PREF_NONE = 2;
}

message SpaceNotificationPreferences {
  // USE_GLOBAL = 0 means fall back to global setting
  optional MentionPreference mentions = 1;
  optional GroupMentionPreference group_mentions = 2;
  optional ThreadPreference threads = 3;
  bool all_messages = 4;  // Per-space only option
}
```

### GraphQL API

```graphql
type Query {
  me {
    notificationPreferences: NotificationPreferences!
    spaceNotificationPreferences(spaceId: ID!): SpaceNotificationPreferences
  }
}

type Mutation {
  updateNotificationPreferences(input: NotificationPreferencesInput!): NotificationPreferences!
  updateSpaceNotificationPreferences(
    spaceId: ID!
    input: SpaceNotificationPreferencesInput!
  ): SpaceNotificationPreferences!
  clearSpaceNotificationPreferences(spaceId: ID!): Boolean!
}

type NotificationPreferences {
  mentions: MentionPreference!
  groupMentions: GroupMentionPreference!
  dms: DMPreference!
  threads: ThreadPreference!
}

# Input types mirror the output types
```

### Notification Decision Flow

When about to send a notification:

```go
func (c *ChattoCore) shouldNotifyUser(
    ctx context.Context,
    userID string,
    spaceID string,
    notificationType NotificationType,
) (bool, error) {
    // 1. Check room mute status (most specific)
    if isMuted, _ := c.IsRoomMuted(ctx, userID, spaceID, roomID); isMuted {
        return false, nil
    }
    
    // 2. Check space-level preference
    spacePref, _ := c.GetSpaceNotificationPrefs(ctx, userID, spaceID)
    if spacePref != nil && spacePref.HasSetting(notificationType) {
        return spacePref.AllowsNotification(notificationType), nil
    }
    
    // 3. Fall back to global preference
    globalPref, _ := c.GetNotificationPrefs(ctx, userID)
    return globalPref.AllowsNotification(notificationType), nil
}
```

## Implementation Tasks

- [ ] Define NotificationPreferences proto messages
- [ ] Implement GetNotificationPrefs / UpdateNotificationPrefs in Core
- [ ] Implement GetSpaceNotificationPrefs / UpdateSpaceNotificationPrefs
- [ ] Add GraphQL types and resolvers
- [ ] Add shouldNotifyUser() check to all notification paths
- [ ] Create /chat/settings/notifications page
- [ ] Create preference select components
- [ ] Create per-space override list UI
- [ ] Add "Mute space" quick action to space menu
- [ ] Add "Notification Settings" link to Notification Center
- [ ] Add E2E test: preferences prevent notifications
- [ ] Add E2E test: space override takes precedence

## Design Decisions

- **Defaults**: Should we default to more or fewer notifications?
  - Proposal: Lean toward more notifications initially, let users turn off
  - Rationale: Users expect notifications; silent failures are worse than noise

- **Sync across devices**: Preferences stored server-side, work on all devices
  - Already handled by storing in KV

- **Migration**: Existing users get default preferences
  - No migration needed; absence of preferences = defaults

## Dependencies

- Mentions feature (chatto-obm2) ✅ Complete
- DM Notifications (for DM preference)
- Thread Notifications (for thread preference)
- Room Muting (for room-level override)

## Future Considerations

- **Notification delivery methods**: Email, push, in-app only
- **Quiet hours**: Don't disturb during specified times
- **Notification sounds**: Different sounds for different types
- **Email digest**: Summary email instead of real-time