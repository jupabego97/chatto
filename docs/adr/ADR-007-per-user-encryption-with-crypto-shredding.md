# ADR-007: Per-User Encryption Keys with Crypto-Shredding for GDPR

**Date:** 2026-03-01

## Context

GDPR's "right to erasure" requires that a user's data can be effectively deleted on request. In a chat application, a user's messages are spread across many streams and rooms. Finding and deleting every message is slow, error-prone, and may leave fragments in backups or replicas.

An alternative is **crypto-shredding**: encrypt each user's messages with a key unique to them, and delete the key when erasure is requested. The encrypted data becomes unreadable without the key, achieving the same practical effect as deletion.

## Decision

Use per-user encryption with crypto-shredding:

- **Algorithm**: ChaCha20-Poly1305 (AEAD). 32-byte keys, 12-byte random nonces prepended to ciphertext.
- **Per-user keys**: Each user has their own encryption key stored in a dedicated `ENCRYPTION_KEYS` KV bucket.
- **Key isolation**: The encryption key bucket is explicitly excluded from `chatto backup`. Backups contain only encrypted data, never the keys to read it.
- **Erasure = key deletion**: When a user requests deletion, their encryption key is removed from the KV bucket. All their messages across all streams become permanently unreadable.
- **KMS service boundary**: Encryption operations (`encrypt`, `decrypt`, `deleteKey`) go through a dedicated KMS service interface. The default implementation is in-process; it can be extracted to a standalone service for high-security deployments.

## Consequences

- **Fast, reliable erasure**: Deleting one KV key renders all of a user's messages unreadable. No need to scan and purge individual messages from streams.
- **Backup safety**: Since keys are excluded from backups, restoring a backup does not restore the ability to read deleted users' messages.
- **No content indexing**: Encrypted message bodies cannot be indexed for full-text search on the server. Search features must either work on metadata or require client-side decryption.
- **Key loss is permanent**: If the KMS loses a user's key (outside of intentional deletion), their messages are gone. The KV bucket must be treated as critical data.
- **Per-message overhead**: Each message has a 12-byte nonce prepended plus the Poly1305 authentication tag. Negligible for chat messages.
- **Future extensibility**: The KMS interface can be adapted to external key management (HashiCorp Vault, AWS KMS, HSM) without changing application code.
