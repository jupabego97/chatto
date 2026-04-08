---
# chatto-gfjj
title: Expose edit window duration from API
status: todo
type: task
created_at: 2026-04-06T11:02:33Z
updated_at: 2026-04-06T11:02:33Z
parent: chatto-p1pf
---

Frontend hardcodes EDIT_WINDOW_MS = 3 * 60 * 60 * 1000 with comment 'matches backend: 3 hours'. Should be API-provided (e.g. on Room or instance config) so client and server stay in sync.
