---
# chatto-bqxe
title: Store MessageDeletedEvent in JetStream for audit trail
status: draft
type: feature
created_at: 2026-01-24T16:40:54Z
updated_at: 2026-01-24T16:40:54Z
---

Currently MessageDeletedEvent is live-only (published to NATS Core, not stored in JetStream). This means there's no audit trail of who deleted what message and when. For compliance and moderation, consider storing delete events in the space event stream alongside other events.