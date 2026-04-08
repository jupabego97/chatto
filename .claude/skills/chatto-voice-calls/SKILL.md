---
name: chatto-voice-calls
description: "Voice call architecture using LiveKit for WebRTC audio. Covers token generation, call state tracking via NATS KV, webhook handling, frontend LiveKit Room integration, audio track attachment, and E2E testing."
---

# Voice Calls (LiveKit)

## Architecture Overview

Voice calls use LiveKit for WebRTC audio. LiveKit runs as an external service -- Chatto handles token generation, server-side call state tracking via NATS KV, and the frontend manages the LiveKit Room connection.

- **Server-side call state** -- LiveKit webhooks notify Chatto of participant join/leave events. Call state is stored in a memory-backed NATS KV bucket (`CALL_STATE`). This is the source of truth for who's in a call.
- **Webhook-driven events** -- join/leave events are published by the **server** (triggered by LiveKit webhooks), not by clients. This means if a client crashes, the leave event still fires when LiveKit detects the disconnection.
- **Graceful degradation** -- all voice APIs return `null`/empty when LiveKit isn't configured. The frontend hides call UI accordingly.

```
LiveKit Server
    | webhook POST
    v
Chatto HTTP Server (/webhooks/livekit)
    | HMAC-validated via livekit/protocol/webhook
    v
Core: update CALL_STATE KV + publish NATS live events
    |
    |-->  CALL_STATE KV (memory-backed): {spaceId}.{roomId} -> JSON participant list
    |
    `-->  NATS live events: CallParticipantJoined / CallParticipantLeft
           |
           v
    Frontend (subscription handler in RoomList.svelte)
```

## Configuration

```toml
[livekit]
enabled = true
url = "wss://livekit.example.com"  # ws:// for local dev
api_key = "your-key"
api_secret = "your-secret"
webhook_url = ""  # defaults to {webserver.url}/webhooks/livekit
```

Environment variables: `CHATTO_LIVEKIT_ENABLED`, `CHATTO_LIVEKIT_URL`, `CHATTO_LIVEKIT_API_KEY`, `CHATTO_LIVEKIT_API_SECRET`, `CHATTO_LIVEKIT_WEBHOOK_URL`.

When enabled, URL, API key, and API secret are required. `webhook_url` defaults to `{webserver.url}/webhooks/livekit`.

## Backend

### Key Files

| File | Purpose |
|------|---------|
| `cli/internal/core/voice.go` | Token generation, call state KV methods, NATS event publishing |
| `cli/internal/graph/voice.resolvers.go` | GraphQL resolvers (auth checks before calling core) |
| `cli/internal/graph/voice.graphqls` | Schema: `voiceCallToken`, `activeCallRoomIds`, `callParticipants` |
| `cli/internal/http_server/webhooks.go` | LiveKit webhook HTTP endpoint with HMAC validation |
| `cli/internal/config/config.go` | `LiveKitConfig` struct and validation |

### Room Naming

LiveKit rooms use deterministic names: `{spaceID}_{roomID}`. `ParseLiveKitRoomName` splits on the first underscore to extract both IDs.

### Token Generation

- Uses `livekit/protocol/auth` to create JWTs with `VideoGrant` (room join permission)
- Tokens are valid for 24 hours
- Identity = user ID, Name = display name
- Metadata = JSON with `login` and `avatarUrl`

### Call State (NATS KV)

Call state is tracked in a global `CALL_STATE` KV bucket with `MemoryStorage`. Call state is ephemeral -- if the server restarts, LiveKit will re-send webhooks as participants reconnect.

- **Key format:** `{spaceId}.{roomId}`
- **Value:** JSON `{"participants": [{"userId", "displayName", "login", "avatarUrl", "joinedAt"}]}`
- **`GetCallParticipants`** -- reads a single KV key
- **`GetActiveCallRoomIDs`** -- lists keys with prefix `{spaceId}.`, extracts room IDs

### Webhook Handler

`POST /webhooks/livekit` receives events from LiveKit, validated via HMAC (`webhook.ReceiveWebhookEvent`):

- `participant_joined` -> add to KV, publish join live event
- `participant_left` -> remove from KV, publish leave live event
- `room_finished` -> publish leave events for remaining participants, delete KV key

### NATS Subjects

Call events are **live events** (direct publish, not stored in JetStream):

| Subject | Event |
|---------|-------|
| `live.space.{spaceId}.room.{roomId}.call_joined` | Participant joined |
| `live.space.{spaceId}.room.{roomId}.call_left` | Participant left |

### GraphQL Authorization

| Operation | Requirement |
|-----------|-------------|
| `voiceCallToken` | Room membership |
| `activeCallRoomIds` | Space membership |
| `callParticipants` | Room membership |

## Frontend

### Key Files

| File | Purpose |
|------|---------|
| `frontend/src/lib/state/voiceCall.svelte.ts` | Singleton state manager wrapping LiveKit Room |
| `frontend/src/lib/state/activeCallRooms.svelte.ts` | Tracks which rooms have active calls |
| `frontend/src/lib/components/voice/VoiceCallButton.svelte` | Join call button in room header |
| `frontend/src/lib/components/voice/VoiceCallPanel.svelte` | In-call UI (participant avatars with status rings, mute, device selector, hang up) |

### Critical: Audio Track Attachment

LiveKit delivers audio data over WebRTC, but **the browser won't play it until the track is attached to an `<audio>` element**. You must handle `RoomEvent.TrackSubscribed` and call `track.attach()`:

```typescript
room.on(RoomEvent.TrackSubscribed, (track, _publication) => {
  if (track.kind === Track.Kind.Audio) {
    track.attach(); // Creates <audio> element and starts playback
  }
});
```

Without this, everything *appears* to work (participants visible, audio levels animating) -- but nobody hears anything.

**Cleanup is equally important** -- call `track.detach()` on `TrackUnsubscribed` and when leaving a call.

### Audio Level Polling

`ActiveSpeakersChanged` fires at ~100ms, which is too coarse for smooth visual feedback. Audio levels are polled at 60ms via `setInterval`, with a change-detection guard to avoid unnecessary Svelte reactivity.

### Call Flow

1. User clicks phone icon -> frontend queries `voiceCallToken`
2. `voiceCallState.join()` creates LiveKit `Room` and connects with token
3. Remote audio tracks arrive via `TrackSubscribed` -> `track.attach()` plays audio
4. LiveKit server fires `participant_joined` webhook -> Chatto updates KV and publishes live event -> other users see headphone icon
5. User hangs up -> LiveKit disconnect -> track cleanup
6. LiveKit fires `participant_left` webhook -> Chatto updates KV and publishes live event -> headphone icon disappears

Note: The frontend does **not** publish join/leave events via mutations. All signaling flows through LiveKit webhooks.

## E2E Testing

Tests configure LiveKit credentials via the `serverOptions` fixture without a real LiveKit server:

```typescript
test.use({
  serverOptions: {
    env: {
      CHATTO_LIVEKIT_ENABLED: 'true',
      CHATTO_LIVEKIT_URL: 'ws://localhost:7880',
      CHATTO_LIVEKIT_API_KEY: 'devkey',
      CHATTO_LIVEKIT_API_SECRET: 'secret'
    }
  }
});
```

Token generation is pure JWT signing -- no LiveKit server needed. Tests verify:
- Token structure (3-part JWT)
- Authorization (room membership required)
- UI visibility (call button in rooms and DMs)
- API responses (`activeCallRoomIds`, `callParticipants`, `instance.livekitUrl`)

For testing real-time call events, test-only webhook endpoints (`/webhooks/test/call-join` and `/webhooks/test/call-leave`) bypass HMAC validation and call core methods directly. These are only available in builds with `-tags test_endpoints`.

Actual WebRTC audio cannot be tested in CI. Manual testing requires two browser windows with different users in the same room.
