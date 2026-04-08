---
# chatto-dahk
title: Denormalize reply preview data
status: todo
type: task
created_at: 2026-04-06T11:02:33Z
updated_at: 2026-04-06T11:02:33Z
parent: chatto-p1pf
---

Every message with inReplyTo fires a separate roomEventByEventId query to fetch the reply preview. 50 replies = 50 queries. Consider adding an inline inReplyToPreview { body, actorDisplayName } field on MessagePostedEvent.
