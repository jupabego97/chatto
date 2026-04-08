---
# chatto-awg7
title: 'Frontend: Room call UI components'
status: todo
type: task
priority: normal
created_at: 2026-01-05T20:40:05Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzz
parent: chatto-or7h
blocking:
    - chatto-0uuj
---

Build UI components for room-wide voice calls.

## Components

### RoomCallPanel.svelte
Side panel showing call participants (replaces member list when in call):
- List of participants with avatars
- Mute/video indicators per participant
- Audio visualizer (optional, nice-to-have)
- Leave call button

### RoomCallBanner.svelte
Banner shown when call is active but user hasn't joined:
- "Voice call in progress - Alice, Bob, +1 more"
- [Join Call] button
- Dismissible (per session)

### RoomCallControls.svelte
Bottom bar controls when in call:
- Mute/unmute toggle
- Video on/off toggle (if we support video)
- Leave call button
- Participant count indicator

### StartCallButton.svelte
Button in room header/input area:
- Shows "Start Call 🎤" when no active call
- Shows "Join Call 🎤" when call is active
- Disabled state when at capacity (4 participants)

## Layout Integration

Modify room layout to:
- Show RoomCallPanel on right side when user is in call
- Show RoomCallBanner above message list when call active but not joined
- Add StartCallButton to room header

## Todo

- [ ] Create RoomCallPanel component
- [ ] Create RoomCallBanner component  
- [ ] Create RoomCallControls component
- [ ] Create StartCallButton component
- [ ] Integrate into room layout
- [ ] Add animations for join/leave
- [ ] Handle "room full" state (4 participants)
