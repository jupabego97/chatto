---
# chatto-zea6
title: Thread participant tracking limited to 50 users
status: draft
type: feature
created_at: 2026-01-24T16:40:54Z
updated_at: 2026-01-24T16:40:54Z
---

Thread metadata in SPACE_{spaceId}_THREADS has a hard-coded cap of 50 participants. Threads with more than 50 unique participants lose newer participants from the tracked list. Consider: expanding the cap, switching to paginated queries, or accepting this as a reasonable limit with documentation.