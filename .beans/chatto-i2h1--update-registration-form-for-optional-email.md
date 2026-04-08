---
# chatto-i2h1
title: Update registration form for optional email
status: draft
type: task
priority: normal
created_at: 2026-01-06T08:02:13Z
updated_at: 2026-01-06T21:13:01Z
---

Update frontend registration form to conditionally require email based on instance config.

## Changes

### Fetch registration config
- Query `registrationConfig { requireEmail }` on registration page load
- Or expose via a public config endpoint

### Registration form
- If `requireEmail` is true: show email field as required
- If `requireEmail` is false: show email field as optional with helper text
- Validation should match config

### Post-registration messaging
- If email was provided: "Check your email to verify your address"
- Explain that some features require email verification

### Location
- `frontend/src/routes/register/+page.svelte` (or similar)

## Notes
- Form already exists, just needs email field to be conditional
- Consider showing what verified users can do vs unverified