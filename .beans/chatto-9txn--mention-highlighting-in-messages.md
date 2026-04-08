---
# chatto-9txn
title: '@mention highlighting in messages'
status: draft
type: feature
priority: normal
created_at: 2026-01-21T18:20:51Z
updated_at: 2026-01-21T18:21:23Z
---

Enhance the display of @mentions in messages with visual highlighting and interactive hover cards.

## Requirements

### Visual Highlighting
- @mentions should be visually distinct from regular text (e.g., colored, background highlight)
- Mentions of the current user should be extra-emphasized (e.g., different color or "ping" style)
- Mentions should be clickable

### Hover Card Integration
- Hovering over a mention shows the user profile hover card (from chatto-oqhm)
- Same behavior as hovering over avatars/names elsewhere

### Click Behavior
- Clicking a mention navigates to the user's profile page

## Implementation Notes
- Extend the existing Markdown renderer to detect and style @mentions
- Mentions are already parsed in the backend - leverage existing mention data
- Coordinate with the hover card component for consistent UX

## Dependencies
- chatto-oqhm (User profile hover cards)

## Tasks

- [ ] Style @mentions distinctly in the Markdown renderer
- [ ] Add special styling for "self mentions" (when current user is mentioned)
- [ ] Integrate hover card triggers on mention elements
- [ ] Add click-to-profile navigation on mentions