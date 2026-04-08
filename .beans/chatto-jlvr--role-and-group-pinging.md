---
# chatto-jlvr
title: Role and Group Pinging
status: draft
type: feature
created_at: 2026-01-20T11:49:01Z
updated_at: 2026-01-20T11:49:01Z
parent: chatto-bq1a
---

Allow users to ping groups of people using special @mentions like @admin, @everyone, @here.

## Overview

Beyond individual @username mentions, users often need to notify groups of people. This feature adds support for special mention targets that expand to multiple recipients.

## Mention Types

### @everyone
- Notifies ALL members of the space
- High-impact, should be restricted by permission
- Use case: Important announcements that everyone must see

### @here
- Notifies members who are currently online/active in the space
- Lower impact than @everyone (only bothers active users)
- Use case: Quick questions, looking for immediate help

### @channel / @room
- Notifies all members of the current room
- Similar to @everyone but scoped to room membership
- Use case: Room-specific announcements

### @role (e.g., @admin, @moderator)
- Notifies all users with a specific role in the space
- Requires role system integration
- Use case: Escalations, role-specific discussions

## Permission Model

Each mention type should have an associated permission:

| Mention | Permission | Default |
|---------|------------|---------|
| @everyone | `mention.everyone` | Admin only |
| @here | `mention.here` | All members |
| @channel | `mention.channel` | All members |
| @{role} | `mention.role` | Role-dependent |

Permissions should be configurable per-space via role settings.

## Technical Design

### Mention Resolution

```go
// ExtractGroupMentions returns special mentions from message body
func ExtractGroupMentions(body string) []GroupMention {
    // Match @everyone, @here, @channel
    // Return structured list of group types
}

// ResolveGroupMention returns user IDs for a group mention
func (c *ChattoCore) ResolveGroupMention(
    ctx context.Context,
    spaceID string,
    mentionType GroupMentionType,
) ([]string, error) {
    switch mentionType {
    case GroupMentionEveryone:
        return c.GetAllSpaceMemberIDs(ctx, spaceID)
    case GroupMentionHere:
        return c.GetOnlineSpaceMemberIDs(ctx, spaceID)
    case GroupMentionChannel:
        return c.GetRoomMemberIDs(ctx, spaceID, roomID)
    case GroupMentionRole:
        return c.GetMembersWithRole(ctx, spaceID, roleID)
    }
}
```

### Permission Check

Before publishing notifications for a group mention:

```go
// Check if actor can use this mention type
canUse, err := core.CanUseMention(ctx, actorID, spaceID, mentionType)
if !canUse {
    // Option 1: Silently ignore the mention (parse but don't notify)
    // Option 2: Return error to user
    // Decision needed
}
```

### Notification Volume Limiting

@everyone in a large space could generate thousands of notifications. Mitigations:
- Rate limit: Max N @everyone mentions per user per hour
- Cooldown: Space-wide cooldown after @everyone (e.g., 5 min)
- Warning: Show confirmation modal before posting @everyone
- Admin override: Admins bypass limits

### Database/Storage Considerations

For @here (online users), need efficient way to query presence:
- Existing presence is stored per-space: `SPACE_{spaceId}_PRESENCE`
- Add helper: `GetOnlineUserIDs(ctx, spaceID) []string`

For @role, need role membership lookup:
- Roles are stored in `SPACE_{spaceId}` bucket
- Member roles in `member.{userId}` entries
- May need index: `role_members.{roleId}` → list of user IDs

### GraphQL Changes

```graphql
# Extend the message posting input
input PostMessageInput {
  # ... existing fields
  # No changes needed - mentions are parsed from body
}

# Permission query for UI hints
type Space {
  viewerCanMentionEveryone: Boolean!
  viewerCanMentionHere: Boolean!
  viewerCanMentionRole(roleId: ID!): Boolean!
}

# Frontend can use these to show/hide autocomplete options
```

### Frontend Autocomplete Integration

When user types `@`:
1. Show individual users (existing)
2. Add special entries: "@everyone", "@here", "@channel"
3. Add role entries: "@admin", "@moderator", etc.
4. Filter based on viewer permissions (don't show @everyone if can't use it)
5. Visual distinction for group mentions (icon, color)

### Notification Deduplication

If user is both @mentioned individually AND part of a group mention:
- Only notify once
- Dedupe: Collect all recipient user IDs into a Set before notifying

## Implementation Tasks

- [ ] Define GroupMentionType enum in proto
- [ ] Implement ExtractGroupMentions() parser
- [ ] Implement ResolveGroupMention() for each type
- [ ] Add mention permissions to role system
- [ ] Add permission checks in PostMessage flow
- [ ] Implement GetOnlineSpaceMemberIDs helper
- [ ] Add role membership index if needed
- [ ] Add viewerCanMention* fields to Space type
- [ ] Update mention autocomplete to show group options
- [ ] Filter autocomplete based on permissions
- [ ] Add rate limiting for @everyone
- [ ] Add confirmation modal for @everyone (frontend)
- [ ] Implement notification deduplication
- [ ] Add E2E tests for group mentions
- [ ] Add E2E tests for permission enforcement

## Design Decisions Needed

- [ ] Should unauthorized @everyone be silently ignored or error?
- [ ] Rate limits: What are reasonable defaults?
- [ ] Should @here include "recently active" (last 5 min) or strict online only?
- [ ] Should @role work across all roles or only some?
- [ ] Add @channel as separate from @room? (Same thing?)

## Security Considerations

- @everyone spam: Rate limiting is essential
- Permission escalation: Ensure role permission checks are robust
- Large spaces: @everyone in 10k+ member space needs careful handling
  - Batch notifications, async processing
  - Consider max recipients limit (notify first N, log rest?)

## Dependencies

- Mentions feature (chatto-obm2) ✅ Complete
- Autocomplete mentions (not yet implemented)
- Role system (partially implemented)