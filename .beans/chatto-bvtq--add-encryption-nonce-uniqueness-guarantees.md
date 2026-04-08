---
# chatto-bvtq
title: Add encryption nonce uniqueness guarantees
status: todo
type: task
priority: low
created_at: 2026-02-16T13:52:32Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzs
parent: chatto-v29q
---

## Problem

The encryption implementation in `encryption.go:39-42` generates 12-byte random nonces using `crypto/rand`. While the probability of collision is astronomically low (~2^-48 after 2^24 messages per user), there's no explicit nonce tracking or collision detection.

## Location

- `cli/internal/encryption/encryption.go` lines 39-42

## Options

### Option A: Document the collision probability (minimal)
Add a comment documenting the birthday bound analysis: with 12-byte nonces and `crypto/rand`, collision probability stays below 2^-32 even with 2^32 messages per user key.

### Option B: Add nonce-misuse resistance (defensive)
Switch to XChaCha20-Poly1305 which uses 24-byte nonces, pushing collision probability to effectively zero. Go's `golang.org/x/crypto/chacha20poly1305` supports this via `NewX()`.

## Notes

- This is a low-priority hardening measure — the current implementation is cryptographically sound for expected message volumes
- Option B would be a breaking change to encrypted message format (acceptable at this stage)
