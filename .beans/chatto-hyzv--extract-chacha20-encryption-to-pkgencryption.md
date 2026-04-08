---
# chatto-hyzv
title: Extract ChaCha20 encryption to pkg/encryption
status: todo
type: task
priority: normal
created_at: 2026-02-28T12:30:06Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzy
parent: chatto-w2dd
---

Extract the pure encryption utilities from cli/internal/encryption into cli/pkg/encryption.

## What to extract

From encryption.go (zero Chatto dependencies):
- Encrypt(key, plaintext) → ciphertext
- Decrypt(key, ciphertext) → plaintext
- GenerateKey() → key

These only depend on stdlib + golang.org/x/crypto/chacha20poly1305.

## What stays or follows

KeyManager in keys.go depends on jetstream.KeyValue for key storage. Options:
1. Extract it too (it only needs a KV interface)
2. Leave it in internal (it's Chatto-specific key lifecycle)

Recommend option 1: extract KeyManager with the KV interface it already uses.

## Tasks
- [ ] Create cli/pkg/encryption package
- [ ] Move Encrypt, Decrypt, GenerateKey functions
- [ ] Move or adapt KeyManager
- [ ] Update imports in cli/internal/core/
- [ ] Verify tests pass
