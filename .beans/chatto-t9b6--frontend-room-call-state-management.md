---
# chatto-t9b6
title: 'Frontend: Room call state management'
status: todo
type: task
priority: normal
created_at: 2026-01-05T20:39:55Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzy
parent: chatto-or7h
blocking:
    - chatto-awg7
---

Rework call state management for room-wide mesh calls.

## Changes from 1:1 Spike

The current callState.svelte.ts is designed for 1:1 calls. Room calls need:
- Multiple RTCPeerConnection instances (one per other participant)
- Dynamic connection management as participants join/leave
- Local media stream shared across all connections

## State Structure

```typescript
type RoomCallState = {
  status: 'idle' | 'joining' | 'connected' | 'error';
  roomId: string | null;
  participants: Map<string, {
    userId: string;
    userName: string;
    isMuted: boolean;
    hasVideo: boolean;
    peerConnection: RTCPeerConnection | null;
    remoteStream: MediaStream | null;
  }>;
  localStream: MediaStream | null;
  isMuted: boolean;
  hasVideo: boolean;
};
```

## Mesh Connection Logic

When joining:
1. Get local media stream
2. For each existing participant, create offer and send via mutation
3. Listen for answers and ICE candidates

When new participant joins:
1. Receive their offer via live event
2. Create answer
3. Exchange ICE candidates

When participant leaves:
1. Close their peer connection
2. Remove from participants map

## Signaling

Reuse existing call signaling events but scoped to room:
- CallOffer, CallAnswer, IceCandidate (already in schema)
- Add roomId to distinguish room calls from 1:1

## Todo

- [ ] Create roomCallState.svelte.ts (separate from 1:1 callState)
- [ ] Implement mesh connection manager
- [ ] Handle dynamic participant join/leave
- [ ] Add local media stream management
- [ ] Wire up to spaceLiveEventBus
- [ ] Add error handling and reconnection logic
