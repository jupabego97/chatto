---
# chatto-vjvc
title: Add encryption stats to admin panel
status: draft
type: task
tags:
    - encryption
created_at: 2026-01-01T12:22:27Z
updated_at: 2026-01-01T12:22:27Z
parent: chatto-vjkr
---

Show encryption-related statistics in the admin panel so operators can understand the encryption state of their instance.

## Proposed Stats
- Encryption enabled: yes/no
- Total user keys: count
- Users with keys: count (vs total users)
- Encrypted messages estimate (could sample or track)

## Location
Add to existing Admin System page or create dedicated Encryption section.

## Notes
This is informational only - no key management actions from admin UI (security boundary).