// Package encryption provides server-side encryption for message bodies
// using ChaCha20-Poly1305 AEAD with per-user keys.
package encryption

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// KeySize is the size of ChaCha20-Poly1305 keys (256 bits).
	KeySize = chacha20poly1305.KeySize // 32 bytes

	// NonceSize is the size of the nonce (96 bits).
	NonceSize = chacha20poly1305.NonceSize // 12 bytes
)

// EncryptedData holds the result of an encryption operation.
type EncryptedData struct {
	Ciphertext []byte
	Nonce      []byte
}

// Encrypt encrypts plaintext using ChaCha20-Poly1305 AEAD.
// Returns the ciphertext (with auth tag) and nonce, or an error.
func Encrypt(key, plaintext []byte) (*EncryptedData, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidKeySize, KeySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD cipher: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt with AEAD (authenticates ciphertext)
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	return &EncryptedData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// Decrypt decrypts ciphertext using ChaCha20-Poly1305 AEAD.
// Returns the plaintext or an error (including authentication failure).
func Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidKeySize, KeySize, len(key))
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidNonceSize, NonceSize, len(nonce))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// GenerateKey generates a cryptographically secure random key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}
