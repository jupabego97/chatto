# Roles & Permissions (RBAC)

## Overview

Chatto uses role-based access control at two levels: **instance-level** (global admin roles) and **space-level** (per-space member roles). Permissions can be further overridden at the **room level** for granular control.

## Space Roles

Every space has four built-in roles (highest to lowest rank):

- **Owner** — Full control over the space, including role management
- **Admin** — Full space permissions except role management
- **Moderator** — Room management, member removal, message moderation
- **Everyone** — Implicit role for all space members. Cannot be assigned or revoked.

Space admins can create **custom roles** that sit between the built-in roles in the hierarchy. Custom roles can be reordered via drag-and-drop.

## Instance Roles

Instance-level roles control global permissions (admin access, DM access, user management). Built-in roles:

- **Instance Owner** — Full instance control
- **Instance Admin** — All instance-level features
- **Instance Moderator** — View-only access
- **Everyone** — Implicit for all authenticated users

## Permission Resolution

Permissions follow a **hierarchy-wins** model:

1. Roles are checked in rank order (highest rank first).
2. The first explicit grant or deny found wins.
3. This means denying a permission on the `everyone` role does NOT block higher-ranked roles.

For example: if `everyone` is denied `message.post` but `admin` is granted it, admins can still post. This enables patterns like read-only announcement channels where only certain roles can post.

## Room-Level Overrides

Space admins can override any permission for any role in a specific room:

- **Grant**: Allow a permission that's denied at the space level
- **Deny**: Block a permission that's granted at the space level
- **Clear**: Remove the override, falling back to the space default

Scope cascade: room > space > instance (more specific scopes win).

## Instance Role Space Permissions

Space admins can also grant or deny space-level permissions to instance roles. This enables patterns like "only users with the `instance:staff` role can create rooms in this space."

## Role Management

- Creating, editing, and deleting roles requires the `role.manage` permission.
- Assigning roles to users requires the `role.assign` permission.
- Users cannot assign or revoke roles equal to or higher than their own rank.
- System roles cannot be deleted. Custom roles can be deleted, which cascades to remove all assignments and permission grants.
