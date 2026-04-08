---
# chatto-0uuj
title: E2E tests for room-wide calls
status: todo
type: task
priority: normal
created_at: 2026-01-05T20:40:18Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzw
parent: chatto-or7h
---

Add e2e tests for room-wide voice call functionality.

## Test Scenarios

### Basic Flow
- [ ] User can start a call in a room
- [ ] Room.activeCall shows call state after starting
- [ ] Second user sees "call in progress" banner
- [ ] Second user can join the call
- [ ] Both users see each other in participant list
- [ ] User can leave call (call continues for others)
- [ ] Last user leaving ends the call

### State Updates
- [ ] Mute state change is visible to other participants
- [ ] Video state change is visible to other participants
- [ ] Participant join/leave updates in real-time

### Edge Cases
- [ ] Cannot start call if one already exists (idempotent join)
- [ ] Call at capacity shows appropriate UI
- [ ] Disconnected user is removed from call
- [ ] Entering room shows existing call state

## Notes

- Use multiple browser contexts (like existing voice-calls.test.ts)
- Mock or skip actual WebRTC media (focus on signaling/state)
- May need longer timeouts for multi-party setup

## Todo

- [ ] Test: start and join room call
- [ ] Test: leave call without ending it
- [ ] Test: last user leaving ends call
- [ ] Test: call state visible on room entry
- [ ] Test: participant list updates in real-time
