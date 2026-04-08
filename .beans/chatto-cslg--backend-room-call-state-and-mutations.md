---
# chatto-cslg
title: 'Backend: Room call state and mutations'
status: todo
type: task
priority: normal
created_at: 2026-01-05T20:39:34Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzy
parent: chatto-or7h
blocking:
    - chatto-t9b6
---

Add backend support for room-wide calls.

## KV State

Add `SPACE_{id}_ROOM_CALLS` KV bucket:
- Key: `{roomId}`
- Value: `{ startedAt, startedBy, participants: [{ userId, joinedAt, isMuted, hasVideo }] }`

## GraphQL Schema

```graphql
type RoomCall {
  roomId: ID!
  startedAt: Time!
  startedBy: User!
  participants: [CallParticipant!]!
}

type CallParticipant {
  user: User!
  joinedAt: Time!
  isMuted: Boolean!
  hasVideo: Boolean!
}

extend type Room {
  activeCall: RoomCall
}

type Mutation {
  startOrJoinRoomCall(roomId: ID!): RoomCall!
  leaveRoomCall(roomId: ID!): Boolean!
  updateMyCallState(roomId: ID!, isMuted: Boolean, hasVideo: Boolean): CallParticipant!
}
```

## Live Events

- RoomCallStarted { roomId, startedBy }
- RoomCallParticipantJoined { roomId, userId }
- RoomCallParticipantLeft { roomId, userId }
- RoomCallParticipantUpdated { roomId, userId, isMuted, hasVideo }
- RoomCallEnded { roomId }

## Authorization

- Must be room member to start/join/leave
- Must be in call to update own state

## Todo

- [ ] Add protobuf types for room call events
- [ ] Add KV bucket initialization in ensureSpaceResources
- [ ] Add GraphQL types and mutations
- [ ] Implement core functions with optimistic locking
- [ ] Add Room.activeCall field resolver
- [ ] Route live events through StreamMySpaceLiveEvents
- [ ] Add tests for concurrent join/leave
