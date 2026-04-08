---
# chatto-hpwo
title: User profile pages
status: draft
type: feature
created_at: 2026-01-21T18:21:04Z
updated_at: 2026-01-21T18:21:04Z
---

Add dedicated profile pages for each user, accessible by clicking on user names, avatars, or member list entries.

## Requirements

### Profile Page Content
- Large avatar
- Display name and username
- Online/offline status
- Bio/about section (if we add this field)
- List of shared spaces (spaces where viewer and profile owner are both members)
- "Send DM" button (if not viewing own profile)
- Edit button (if viewing own profile)

### Navigation
- Route: `/chat/user/{userId}` or `/chat/profile/{userId}`
- Accessible from:
  - Clicking user avatar in messages
  - Clicking user name in messages
  - Clicking entry in member list
  - "View Profile" button in hover card
  - Clicking own avatar in sidebar/header

### Own Profile
- Should also be accessible from `/chat/settings/profile` or similar
- Allow editing display name, avatar, bio

## Implementation Notes

### Frontend
- New route under `/chat/user/[userId]/+page.svelte`
- Query user data via GraphQL
- Handle "user not found" gracefully

### Backend
- May need to add fields like `bio` to User model
- Add `sharedSpaces` resolver to find mutual memberships
- User profiles are public (per authorization model)

## Tasks

### Backend
- [ ] Add `bio` field to User (optional)
- [ ] Add `sharedSpaces(viewerId: ID!)` field to User type
- [ ] Implement `sharedSpaces` resolver

### Frontend  
- [ ] Create `/chat/user/[userId]/+page.svelte` route
- [ ] Create `UserProfile.svelte` component
- [ ] Query and display user data
- [ ] Show shared spaces list
- [ ] Add "Send DM" action button
- [ ] Handle own profile view vs other user view
- [ ] Add click handlers to avatars/names to navigate to profile