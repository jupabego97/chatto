---
# chatto-xsfr
title: Make DM space ID discoverable from API
status: todo
type: task
created_at: 2026-04-06T11:02:33Z
updated_at: 2026-04-06T11:02:33Z
parent: chatto-p1pf
---

Frontend hardcodes DM_SPACE_ID = 'DM' in constants.ts with warning it must match the Go constant. Should be discoverable via the API (e.g. me.dmSpaceId or instance.dmSpaceId).
