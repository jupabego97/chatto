---
# chatto-9ea9
title: Extract KMS as NATS micro service
status: draft
type: feature
priority: normal
tags:
    - encryption
created_at: 2025-12-31T11:10:30Z
updated_at: 2026-01-02T20:46:00Z
parent: chatto-vjkr
---

Move encryption key management from INSTANCE KV to a dedicated NATS micro service.

## Goals
- Keys stored in separate KV bucket (excluded from backups)
- Service exposes encrypt/decrypt/deleteKey operations
- Runs embedded by default, can be extracted to standalone
- Enables future adapters (Vault, AWS KMS, HSM)

## Tasks
- [ ] Create dedicated KV bucket for encryption keys
- [ ] Implement NATS micro service interface
- [ ] Migrate key storage from INSTANCE bucket
- [ ] Update backup command to exclude key bucket
- [ ] Update ChattoCore to use service instead of direct key access
- [ ] Add e2e test for encryption (message content encrypted at rest)

## Dependencies
- Depends on: chatto-6swh (server-side encryption, completed)
- Blocked by: chatto-kxd4 (DeleteUser for GDPR) should be done first

## References
- See .claude/rules/encryption.md for architecture details