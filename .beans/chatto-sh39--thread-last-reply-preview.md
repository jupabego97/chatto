---
# chatto-sh39
title: Thread last reply preview
status: draft
type: feature
priority: normal
tags:
    - threading
    - frontend
created_at: 2026-01-02T18:59:46Z
updated_at: 2026-01-02T20:45:43Z
parent: chatto-d2tk
---

Show a snippet of the last reply and relative time (e.g., '2h ago') in the thread indicator, not just reply count. The backend already stores lastReplyAt in ThreadMetadata; need to add last reply snippet and surface it in the UI.