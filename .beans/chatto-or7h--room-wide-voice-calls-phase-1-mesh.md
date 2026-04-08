---
# chatto-or7h
title: Room-wide voice calls (Phase 1 - Mesh)
status: draft
type: epic
created_at: 2026-01-05T20:39:24Z
updated_at: 2026-01-05T20:39:24Z
---

Room-scoped voice calls where any room member can join an ongoing call. Similar to Discord voice channels or Slack huddles.

## Goals
- One active call per room (max ~4 participants due to mesh topology)
- Persistent call state visible to all room members
- Join/leave freely without ending the call
- Mute/video state per participant

## Non-Goals (Phase 2)
- SFU for scaling beyond 4 participants
- Screen sharing
- Call recording

## Architecture
- KV bucket for call state (participants, start time)
- Live events for real-time UI updates
- Mesh WebRTC topology (each participant connects to all others)

## UI
- Side panel showing participants when call is active
- "Start Call" button in room (becomes "Join Call" when active)
- Banner when call is active but user hasn't joined