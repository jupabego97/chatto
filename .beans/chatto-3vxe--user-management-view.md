---
# chatto-3vxe
title: User management view
status: draft
type: feature
tags:
    - admin
    - frontend
    - backend
created_at: 2025-12-21T21:26:00Z
updated_at: 2025-12-21T21:26:00Z
parent: chatto-s5ot
---

Admin view for managing registered users on the instance.

## Capabilities

| Action | Description |
|--------|-------------|
| List users | Paginated table with search |
| View user details | Profile info, spaces joined, created date |
| Disable user | Prevent login, revoke sessions |
| Enable user | Re-enable disabled user |
| Reset password | Generate password reset (requires email) |
| Force logout | Invalidate all sessions |

## What Admins Cannot Do

- View user's messages
- Join spaces on behalf of user
- See user's private data beyond profile

## Backend Requirements

- GraphQL queries for user listing with pagination/search
- Mutations for disable/enable/force-logout
- New fields on User: `disabled`, `disabled_at`, `disabled_by`

## Checklist

- [ ] Add `disabled` field to User model
- [ ] Add admin GraphQL queries for user listing
- [ ] Add admin mutations (disableUser, enableUser, forceLogout)
- [ ] Create user list page with data table
- [ ] Create user detail view
- [ ] Add disable/enable actions
- [ ] Add force logout action