---
# chatto-ydr9
title: Add string length validation in GraphQL resolvers
status: todo
type: task
priority: normal
created_at: 2026-02-16T13:52:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzw
parent: chatto-v29q
---

## Problem

GraphQL mutation inputs lack server-side string length validation. Malicious or buggy clients can send arbitrarily long strings for space names, room names, message bodies, etc. This data flows through to NATS subjects and JetStream storage without bounds checking.

## Locations

- `cli/internal/graph/schema.resolvers.go` — mutation resolvers for createSpace, createRoom, postMessage, updateSpace, updateRoom
- `cli/internal/graph/schema.graphqls` — input type definitions

## Recommended Approach

1. Define max lengths as constants (e.g., MaxSpaceName=100, MaxRoomName=100, MaxMessageBody=10000, MaxDisplayName=50)
2. Add validation at the top of each mutation resolver, before calling core
3. Return clear GraphQL errors with the field name and limit
4. Consider a shared `validateStringLength(field, value, max)` helper

## Notes

- This is a security boundary concern — validate at the API layer per project architecture
- NATS subjects have a max length of 255 bytes; extremely long space/room names could cause subject construction failures
