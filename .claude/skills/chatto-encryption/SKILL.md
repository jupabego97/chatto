---
name: chatto-encryption
description: "Encryption architecture for message body encryption and GDPR crypto-shredding. Covers ChaCha20-Poly1305, per-user keys, KMS service design, and key storage."
---

# Encryption Architecture

## Core Principles

1. **Per-user keys** - Each user has their own encryption key. This enables GDPR-compliant crypto-shredding: deleting a user's key renders all their messages unreadable.
2. **Keys are separate from data** - Key storage is excluded from `chatto backup` so backups contain only encrypted data.
3. **Service boundary** - Encryption operations go through a dedicated KMS service, not direct key access.

## Algorithm

- **ChaCha20-Poly1305** (AEAD)
- 32-byte keys, 12-byte random nonces
- Nonce stored alongside ciphertext (prepended)

## KMS Service

A NATS micro service manages user keys and encryption operations:

```
encrypt(userId, plaintext) -> ciphertext
decrypt(userId, ciphertext) -> plaintext
deleteKey(userId) -> void  // For GDPR deletion
```

Design goals:
- **Embeddable**: Runs in-process by default (single executable goal)
- **Extractable**: Can run standalone for high-security deployments
- **Extensible**: Interface allows future adapters (Vault, AWS KMS, HSM)

## Storage

- Keys stored in dedicated NATS KV bucket (not `INSTANCE`)
- Key bucket excluded from backup commands
- JetStream KV provides memory-backed caching automatically

## GDPR Compliance

When a user requests deletion:
1. Delete their encryption key from KMS
2. All their encrypted messages become unreadable
3. Message records can remain (or be purged) - content is cryptographically destroyed

This is "crypto-shredding" - faster and more reliable than finding/deleting every message across streams.

## Current State

Message body encryption is implemented with keys in a dedicated `ENCRYPTION_KEYS` KV bucket (excluded from backups). The full KMS micro service extraction is planned work (see `chatto-9ea9`).
