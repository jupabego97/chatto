---
# chatto-a4il
title: Independent echo deletion — deleting echo should not delete thread original
status: todo
type: bug
priority: high
created_at: 2026-04-05T20:24:58Z
updated_at: 2026-04-05T20:26:02Z
---

Currently, deleting an echo also deletes the original thread reply because they share the same message body in KV. Deleting either one removes the shared body, making both show as deleted. The echo should be independently deletable without affecting the original thread reply.

## Context

Echo events and their original thread replies share the same `messageBodyId` — a single KV entry in the bodies bucket. Deleting either event calls `DeleteMessage` which removes that KV entry, making both the echo and the original show as deleted.

## Desired behavior

- Deleting an echo should only hide the echo from the main room timeline
- The original thread reply should remain intact with its body readable
- Deleting the original thread reply should still make the echo show as deleted (the echo is a view of the original's content)

## Possible approaches

1. **Delete the echo event from JetStream** instead of deleting the body. The echo event would be removed from the stream, so it no longer appears in room history. The original's body stays intact.
2. **Give echoes their own body copy** instead of sharing. This decouples deletion but doubles storage and breaks edit propagation (edits would need to update both copies).
3. **Soft-delete via a flag** on the echo event rather than removing the body. Add a "hidden" or "deleted" marker that the frontend respects, without touching the shared body.

Approach 1 is not feasible — JetStream streams are append-only, individual messages cannot be removed. Approach 3 is likely the right path: a soft-delete marker in KV (e.g., `echo_hidden.{eventId}`) that the resolver checks when loading room events. The echo would be filtered out during loading, similar to how deleted message bodies are handled.
