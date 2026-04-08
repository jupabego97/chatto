---
# chatto-oqhm
title: User profile hover cards
status: draft
type: feature
priority: normal
created_at: 2026-01-21T18:20:41Z
updated_at: 2026-01-21T18:21:27Z
blocking:
    - chatto-9txn
---

Implement floating user profile cards that appear when hovering over:
- User avatars in messages
- User names in messages  
- User entries in the room member list

## Requirements

### Hover Card Content
- User avatar (larger)
- Display name
- Username/login
- Online/offline status indicator
- "View Profile" button/link
- "Send DM" button (if not self)

### Behavior
- Appear after a short delay (~300ms) to prevent accidental triggers
- Stay visible while mouse is over the card
- Dismiss when mouse leaves both trigger and card
- Position intelligently (flip if near edge of viewport)

### Implementation Notes
- Create a shared `UserHoverCard.svelte` component
- Use Svelte's portal pattern or a global container for positioning
- Consider using a library like Floating UI for positioning logic
- Reuse existing user data from context where possible

## Tasks

- [ ] Create `UserHoverCard.svelte` component
- [ ] Add hover triggers to message avatars
- [ ] Add hover triggers to message author names  
- [ ] Add hover triggers to member list entries
- [ ] Implement smart positioning (viewport edge detection)
- [ ] Add "View Profile" and "Send DM" actions