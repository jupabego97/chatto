---
# chatto-64rr
title: Space and room overview
status: draft
type: feature
tags:
    - admin
    - frontend
    - backend
created_at: 2025-12-21T21:26:00Z
updated_at: 2025-12-21T21:26:00Z
parent: chatto-s5ot
---

Admin view for browsing spaces and rooms on the instance.

## Capabilities

### Spaces
- List all spaces (paginated, searchable)
- View space metadata: name, member count, room count, created date
- View space admins
- Delete space (with confirmation)

### Rooms (per space)
- List rooms in a space
- View room metadata: name, message count, created date
- Delete room (with confirmation)

## Privacy Boundary

- Admins can see metadata (counts, dates, names)
- Admins CANNOT read message content
- Admins CANNOT see who sent which message
- This is intentional and important

## Backend Requirements

- Admin GraphQL queries for space/room listing
- Aggregation queries for counts (member count, message count)
- Delete mutations with proper cleanup

## Checklist

- [ ] Add admin GraphQL queries for space listing
- [ ] Add aggregation resolvers (memberCount, roomCount, messageCount)
- [ ] Add deleteSpace mutation (admin only)
- [ ] Add deleteRoom mutation (admin only)
- [ ] Create space list page
- [ ] Create space detail view with room list
- [ ] Add delete confirmations