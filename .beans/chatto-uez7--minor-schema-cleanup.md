---
# chatto-uez7
title: Minor schema cleanup
status: todo
type: task
priority: low
created_at: 2026-04-06T11:02:42Z
updated_at: 2026-04-06T11:02:42Z
parent: chatto-p1pf
---

Low-priority schema hygiene:
- Add descriptions to VideoProcessingStatus enum values
- Rename TimeFormat.UNSPECIFIED → AUTO
- Fix StreamInfo.created: String! → Time!
- Add separate PresenceStatusInput enum (exclude OFFLINE from input)
- Align avatarUrl/logoUrl to also accept fit parameter
