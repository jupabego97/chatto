---
# chatto-jrzb
title: 'Frontend: Multi-instance voice calls'
status: draft
type: feature
priority: normal
created_at: 2026-03-03T11:49:17Z
updated_at: 2026-03-03T11:49:30Z
parent: chatto-wadw
blocked_by:
    - chatto-46kr
---

Scope voice call functionality to the correct instance. Each instance may have its own LiveKit server.

## Context

Voice calls use LiveKit for WebRTC. The LiveKit URL comes from instance config (`instanceState.livekitUrl`). Call state is tracked per-instance via NATS KV and GraphQL subscriptions.

## Requirements

- [ ] LiveKit URL is per-instance (read from each instance's config)
- [ ] Active call state is per-instance
- [ ] Call participant list is per-instance
- [ ] Only one active call at a time (across all instances) — this is a UX constraint, not a technical one
- [ ] Joining a call on instance B while in a call on instance A prompts to leave the first call
- [ ] Call indicators in the sidebar show which rooms have active calls (per instance)
- [ ] Write tests

## Implementation Notes

The voice call state (`voiceCallState`, `callParticipantsState`, `activeCallRooms`) move into `InstanceState`. The LiveKit Room connection uses the instance-specific LiveKit URL.

### Key files
- `frontend/src/lib/state/voiceCall.svelte.ts`
- `frontend/src/lib/state/callParticipants.svelte.ts`
- `frontend/src/lib/state/activeCallRooms.svelte.ts`

### Blocked by
- Instance-scoped state management (voice state lives in InstanceState)

### Priority
Low — voice calls are a secondary feature. Can be deferred until the core multi-instance flow works.
