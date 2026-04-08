---
# chatto-af1l
title: Fix iOS PWA composer scrolling off screen
status: in-progress
type: bug
created_at: 2026-03-25T17:25:18Z
updated_at: 2026-03-25T17:25:18Z
---

In iOS PWA standalone mode, the chat message composer is pushed below the visible area. Root cause: h-dvh on root layout doesn't account for safe-area padding on the body, making the flex container taller than the body's content box.
