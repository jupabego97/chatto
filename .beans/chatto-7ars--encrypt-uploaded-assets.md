---
# chatto-7ars
title: Encrypt uploaded assets
status: draft
type: feature
created_at: 2026-01-21T18:59:39Z
updated_at: 2026-01-21T18:59:39Z
---

Uploaded attachments are currently stored in plaintext while message bodies are encrypted. This creates an inconsistency in the security model and breaks GDPR crypto-shredding for attachments.

## Current State

**Message bodies** are encrypted using ChaCha20-Poly1305 with per-user keys stored in `ENCRYPTION_KEYS` KV bucket. This enables GDPR-compliant crypto-shredding: deleting a user's key renders all their messages unreadable.

**Uploaded attachments** are stored in plaintext in NATS ObjectStore buckets:
- Instance assets (avatars, logos): `INSTANCE_ASSETS`
- Space attachments (images, files): `SPACE_{spaceId}_ASSETS`

## Security Gap

| Aspect | Messages | Attachments |
|--------|----------|-------------|
| Encrypted | Yes (ChaCha20-Poly1305) | **No** |
| GDPR crypto-shred | Works | Doesn't apply |
| Backup exposure | Safe (encrypted) | **Readable** |
| At-rest protection | Yes | **No** |

## Why Encrypt Assets?

1. **GDPR compliance** - Crypto-shredding should apply to all user content, not just message text
2. **Data-at-rest protection** - Anyone with NATS data file access can read all attachments
3. **Backup safety** - Defeats the purpose of excluding encryption keys from backups
4. **Consistency** - Users expect their uploaded files to have the same protection as their messages

## Design Considerations

### Key Management
- Use the same per-user keys from `ENCRYPTION_KEYS` bucket
- Attachment encrypted with uploader's key
- Same crypto-shredding benefit: delete key → attachments become unreadable

### Performance Tradeoffs
- **CPU overhead**: Encrypt on upload, decrypt on every access
- **Large files**: May need chunked encryption for streaming support
- **Memory**: Can't stream directly from ObjectStore without buffering

### Image Transformation Cache
Current flow: ObjectStore → transform → cache in `ASSET_CACHE`

With encryption:
- Option A: Cache decrypted transforms (faster access, but cache contains plaintext)
- Option B: Re-encrypt cached transforms (consistent protection, more CPU)
- Option C: Don't cache encrypted assets (simpler, but slower)

### Implementation Approach
1. Encrypt on upload in `UploadAttachment()` using uploader's key
2. Store encryption metadata (nonce, key ID) in ObjectStore headers
3. Decrypt on access in asset serving HTTP handler
4. Update image transformation pipeline to work with encrypted source

### Migration
- Existing attachments remain unencrypted (or run background migration)
- New attachments encrypted going forward
- Could add `encrypted` flag to distinguish

## Related Code

- `cli/internal/core/attachments.go` - Upload/storage logic
- `cli/internal/core/assets.go` - Asset serving HTTP handlers
- `cli/internal/encryption/` - ChaCha20-Poly1305 implementation
- `cli/internal/core/rooms.go:1067-1082` - Message encryption pattern to follow