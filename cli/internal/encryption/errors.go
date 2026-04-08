package encryption

import "errors"

var (
	// ErrDecryptionFailed indicates the ciphertext couldn't be decrypted
	// (wrong key, corrupted data, or tampered ciphertext).
	ErrDecryptionFailed = errors.New("decryption failed: invalid key or corrupted data")

	// ErrKeyNotFound indicates no encryption key exists for the requested entity.
	ErrKeyNotFound = errors.New("encryption key not found")

	// ErrInvalidKeySize indicates the provided key has an incorrect size.
	ErrInvalidKeySize = errors.New("invalid key size")

	// ErrInvalidNonceSize indicates the provided nonce has an incorrect size.
	ErrInvalidNonceSize = errors.New("invalid nonce size")
)
