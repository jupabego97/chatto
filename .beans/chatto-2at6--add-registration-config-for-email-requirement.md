---
# chatto-2at6
title: Add registration config for email requirement
status: draft
type: task
priority: normal
created_at: 2026-01-06T08:01:49Z
updated_at: 2026-01-06T21:13:01Z
blocking:
    - chatto-i2h1
---

Add instance configuration option to require email at registration.

## Config

```toml
[registration]
require_email = false  # default: email is optional
```

## Behavior

When `require_email = true`:
- Registration form shows email field as required
- User account is created immediately (no blocking on verification)
- Verification email is sent
- User gets `verified_users` role only after clicking verification link

When `require_email = false`:
- Registration form shows email field as optional
- If email provided, same verification flow as above
- If no email provided, user can add one later in settings

## Changes

### config/config.go
- Add `Registration` struct with `RequireEmail bool`
- Add to main Config struct

### GraphQL
- Expose config to frontend: `Query.registrationConfig: RegistrationConfig!`
- `RegistrationConfig { requireEmail: Boolean! }`

### Registration endpoint
- Validate email presence when config requires it
- Always send verification email when email is provided

## Notes
- This is about the registration form, not account creation
- OIDC bypasses this (email comes from provider)
- Frontend will read this config to show/hide email field