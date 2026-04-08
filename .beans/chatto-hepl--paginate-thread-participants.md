---
# chatto-hepl
title: Paginate thread participants
status: draft
type: task
priority: normal
tags:
    - threading
    - backend
    - performance
created_at: 2026-01-02T18:59:52Z
updated_at: 2026-01-02T20:45:44Z
parent: chatto-d2tk
---

Thread participants are currently capped at 50 in the KV metadata. For very active threads with many participants, consider: pagination in the GraphQL API, 'and N more' indicator in UI, or lazy-loading additional participants on demand.