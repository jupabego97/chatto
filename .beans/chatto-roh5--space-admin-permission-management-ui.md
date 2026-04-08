---
# chatto-roh5
title: Space admin permission management UI
status: todo
type: feature
priority: normal
tags:
    - frontend
    - backend
    - auth
created_at: 2025-12-26T22:14:01Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzz
parent: chatto-0kda
---

Build a simple permission management interface for space admins to assign roles and permissions to users.

## Requirements

- View space members with their current roles
- Assign/remove roles from users
- View what permissions each role grants
- Simple, clear UI that doesn't overwhelm

## Checklist

### Backend
- [ ] GraphQL queries for space members with roles
- [ ] Mutations for role assignment
- [ ] Ensure proper authorization checks

### Frontend
- [ ] Space settings page/section for member management
- [ ] Member list with role display
- [ ] Role assignment UI (dropdown/modal)
- [ ] Permission explanation/tooltips
