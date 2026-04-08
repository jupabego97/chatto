---
# chatto-jh9o
title: Room-level permissions system
status: todo
type: feature
priority: normal
created_at: 2026-01-18T22:35:52Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzy
---

## Overview

Extend the permission system to support room-level permissions, allowing space admins to configure which roles can perform specific actions within individual rooms.

## Proposed Room Permissions

- `room.leave` - Can leave the room (may want to lock members in certain rooms)
- `room.manage` - Can update room settings (name, description, topic)
- `room.delete` - Can delete the room
- `room.invite` - Can invite/add members to the room
- `room.remove` - Can remove members from the room
- `room.post` - Can post messages in the room (for read-only announcement rooms)
- `room.pin` - Can pin messages
- `room.moderate` - Can delete others' messages

## Backend

- Extend permission model to support room-scoped permissions
- Add room permission overrides (role has permission X in room Y)
- Update `Can*` functions to check room-level permissions
- Consider inheritance: space permission as default, room permission as override

## Frontend

- Room settings UI for managing permissions per role
- Visual indicators for restricted rooms (e.g., read-only badge)
- Disable/hide actions user doesn't have permission for

## Design Considerations

- Permission inheritance model: deny-by-default or allow-by-default with overrides?
- Per-room role assignment vs using space roles with room overrides
- UI complexity: keep it simple for common cases, advanced for power users
- Migration path for existing rooms
