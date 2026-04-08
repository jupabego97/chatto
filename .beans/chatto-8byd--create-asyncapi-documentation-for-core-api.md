---
# chatto-8byd
title: Create AsyncAPI documentation for Core API
status: draft
type: feature
created_at: 2026-01-09T09:00:02Z
updated_at: 2026-01-09T09:00:02Z
---

Generate and maintain AsyncAPI specification for the NATS-based Core API. This provides industry-standard documentation for the message-based API layer between GraphQL and Core.

## Goals

- Document all NATS service endpoints (subjects, request/response schemas)
- Enable tooling ecosystem (documentation viewers, client generation)
- Provide clear reference for future native clients (mobile, CLI)

## Approach

1. Use proto files as source of truth for message schemas
2. Generate AsyncAPI spec from proto + subject pattern config
3. Generate human-readable docs (Markdown/HTML) from AsyncAPI
4. Add to CI to keep docs in sync

## Research needed

- Evaluate `protoc-gen-asyncapi` or similar tools
- Determine if manual AsyncAPI + proto import is cleaner
- Consider hosting options (GitHub Pages, embedded in app)

## Example structure

```yaml
asyncapi: '2.6.0'
info:
  title: Chatto Core API
  version: '1.0.0'
channels:
  rooms.{actorId}.create:
    publish:
      message:
        $ref: '#/components/messages/CreateRoomRequest'
    subscribe:
      message:
        $ref: '#/components/messages/CreateRoomResponse'
```

## Services to document

- spaces (create, get, list, join, leave, update)
- rooms (create, get, list, join, leave, members, events, read state)
- messages (post, edit, delete)
- reactions (add, remove, list)
- users (get, list, update)
- permissions (get, set, list-roles)
- presence (get, update)
- dms (find-or-create, list, participants)
- kms (encrypt, decrypt, delete-key)