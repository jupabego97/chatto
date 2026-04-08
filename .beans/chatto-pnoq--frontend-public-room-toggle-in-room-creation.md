---
# chatto-pnoq
title: 'Frontend: public room toggle in room creation'
status: todo
type: task
priority: normal
created_at: 2026-02-10T08:05:54Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzs
parent: chatto-2quz
blocked_by:
    - chatto-j6h3
---

Add a "Public room" option to the room creation flow so admins can designate rooms as public at creation time.

## Implementation

### Create Room modal
- Add a checkbox/toggle: "Make this room public"
- Below it, explanatory text: "Public rooms can be viewed by anyone on the internet without an account. This cannot be changed after creation."
- Only show this option if the user has \`rooms.manage\` permission (or dedicated permission)
- Wire up the \`public\` field in the \`CreateRoomInput\` mutation

### Room settings page
- Show a read-only indicator: "This room is public" or "This room is private"
- Do NOT allow toggling (immutable after creation)
- For public rooms, show the public URL that can be shared

### Room list indicators
- Add a visual indicator (icon or badge) for public rooms in the sidebar and Browse Rooms page
- Something like a globe icon or "Public" chip
- Visible to space members so they know which rooms are publicly accessible

### Admin rooms page
- Show public/private status in the room management table
- The drag-and-drop admin page should display public indicator on room cards

## Key files
- \`frontend/src/lib/components/CreateRoom.svelte\` (or wherever the create room modal lives)
- \`frontend/src/routes/chat/[spaceId]/[roomId]/settings/\` (room settings)
- Sidebar room list components
- Admin rooms page

## Tests
- E2E: admin creates a public room via the modal
- E2E: public room indicator visible in room list
- E2E: room settings shows immutable public status
- E2E: non-admin user does not see the public room toggle
