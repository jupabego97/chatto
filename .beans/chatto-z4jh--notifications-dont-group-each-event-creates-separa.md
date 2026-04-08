---
# chatto-z4jh
title: Notifications don't group - each event creates separate entry
status: draft
type: feature
created_at: 2026-01-24T16:40:54Z
updated_at: 2026-01-24T16:40:54Z
---

Each mention or thread reply creates a separate notification entry in the NOTIFICATIONS KV bucket. In active spaces, this could grow unbounded and overwhelm users. Consider implementing notification aggregation/digest (e.g., '5 new mentions in #general' instead of 5 separate notifications).