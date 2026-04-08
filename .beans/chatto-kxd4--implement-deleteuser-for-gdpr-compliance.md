---
# chatto-kxd4
title: Implement DeleteUser for GDPR compliance
status: todo
type: feature
priority: normal
tags:
    - encryption
created_at: 2025-12-31T00:04:59Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzV
parent: chatto-vjkr
---

Add a `DeleteUser()` function to Core that handles complete user deletion for GDPR compliance.

## Requirements

- Delete the user's encryption key (crypto-shredding - makes all their messages unreadable)
- Anonymize or delete the user record itself
- Ensure frontend gracefully handles deleted/anonymized users (e.g., show "[deleted user]" for author names)
- **E2E test**: Verify that after deleting a user's key, their messages show "[message deleted]"

## Open Questions

- Should we soft-delete (anonymize: clear PII, keep ID for referential integrity) or hard-delete the user record?
- What happens to spaces/rooms the user created or owns?
- Should we provide a grace period before permanent deletion?

## Related

- Builds on chatto-6swh (server-side message encryption with crypto-shredding)
