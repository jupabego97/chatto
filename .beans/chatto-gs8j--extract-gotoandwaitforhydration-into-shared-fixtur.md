---
# chatto-gs8j
title: Extract gotoAndWaitForHydration into shared fixture
status: todo
type: task
created_at: 2026-03-18T15:02:42Z
updated_at: 2026-03-18T15:02:42Z
parent: chatto-qfp3
---

The gotoAndWaitForHydration pattern (wait for WS console log + networkidle) exists only in session-expiration.test.ts. Extract to shared fixture and use across tests that navigate.
