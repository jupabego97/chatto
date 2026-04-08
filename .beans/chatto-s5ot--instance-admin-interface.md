---
# chatto-s5ot
title: Instance Admin Interface
status: draft
type: epic
priority: normal
tags:
    - admin
    - frontend
    - backend
created_at: 2025-12-21T21:25:22Z
updated_at: 2026-01-02T20:45:59Z
parent: chatto-vjkr
---

A dedicated admin interface for instance operators, completely separate from the chat functionality.

## Goals

- Provide instance operators with visibility into their deployment
- Allow runtime configuration of instance settings
- Maintain privacy boundary: admins see metadata, not message content

## What Admins Can See/Do

| Category | Capabilities |
|----------|-------------|
| **Users** | List, search, disable/enable, reset password, force logout |
| **Spaces** | List, view metadata (member count, room count), delete |
| **Rooms** | List per space, metadata only (message count, created date) |
| **System** | NATS streams, storage usage, consumer health |
| **Config** | Runtime-adjustable settings |

## What Admins Cannot Do

- Read message content
- Join rooms they're not members of
- Access private conversations

This is an important trust property for users.

## Technical Approach

- Separate route at `/admin` within the same SPA
- Different layout/navigation from chat UI
- Instance-level admin role (separate from per-space admin)

## Deferred

- **Audit trail**: Logging of admin actions will be addressed later

## Open Questions

- How does someone become an instance admin? (first user? TOML config? init command?)