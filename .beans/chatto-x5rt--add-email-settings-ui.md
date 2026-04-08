---
# chatto-x5rt
title: Add email settings UI
status: draft
type: task
priority: normal
created_at: 2026-01-06T08:02:22Z
updated_at: 2026-01-06T21:13:01Z
---

Add UI for users to manage their email addresses.

## Location

User settings/profile page (or new email settings section)

## Features

### View verified emails
- List all verified emails for current user
- Show which is primary (for future use)
- Option to remove non-primary emails

### Add new email
- Input field for email address
- "Send verification" button
- Shows pending verification status
- Clear messaging: "Verification email sent to X"

### Verification status
- Show if user has `verified_users` status
- Explain benefits of email verification
- Prompt to add email if none verified

## Components

- Email list component
- Add email form
- Pending verification indicator

## GraphQL Operations

- `me { emails hasVerifiedEmail }` - list current emails
- `requestEmailVerification(email: String!)` - start verification

## Notes
- Only show verified emails (pending ones shown as "pending verification")
- Design for multiple emails even if v1 is single email
- Consider resend verification link option