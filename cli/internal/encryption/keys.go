package encryption

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// KeyManager handles encryption key operations.
type KeyManager struct {
	kv jetstream.KeyValue
}

// NewKeyManager creates a new key manager using the provided KV bucket.
// The bucket should be the dedicated ENCRYPTION_KEYS bucket.
func NewKeyManager(kv jetstream.KeyValue) *KeyManager {
	return &KeyManager{kv: kv}
}

// userKeyPath returns the KV key for a user's encryption key.
// Keys are stored directly by user ID since the bucket is dedicated to encryption.
func userKeyPath(userID string) string {
	return "user." + userID
}

// GetUserKey retrieves a user's encryption key.
// Returns nil, nil if no key exists (encryption disabled or crypto-shredded).
func (m *KeyManager) GetUserKey(ctx context.Context, userID string) ([]byte, error) {
	entry, err := m.kv.Get(ctx, userKeyPath(userID))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, nil // No key = messages will be plaintext or unreadable
		}
		return nil, fmt.Errorf("failed to get user encryption key: %w", err)
	}
	return entry.Value(), nil
}

// CreateUserKey generates and stores a new encryption key for a user.
// Uses kv.Create for atomicity (fails if key already exists).
func (m *KeyManager) CreateUserKey(ctx context.Context, userID string) ([]byte, error) {
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	_, err = m.kv.Create(ctx, userKeyPath(userID), key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyExists) {
			// Key already exists, return it
			return m.GetUserKey(ctx, userID)
		}
		return nil, fmt.Errorf("failed to store user encryption key: %w", err)
	}

	return key, nil
}

// GetOrCreateUserKey retrieves an existing key or creates a new one.
func (m *KeyManager) GetOrCreateUserKey(ctx context.Context, userID string) ([]byte, error) {
	key, err := m.GetUserKey(ctx, userID)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return key, nil
	}
	return m.CreateUserKey(ctx, userID)
}

// DeleteUserKey permanently deletes a user's encryption key (crypto-shredding).
// All messages encrypted with this key become permanently unreadable.
func (m *KeyManager) DeleteUserKey(ctx context.Context, userID string) error {
	err := m.kv.Delete(ctx, userKeyPath(userID))
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		return fmt.Errorf("failed to delete user encryption key: %w", err)
	}
	return nil
}

// UserKeyExists checks if a user has an encryption key.
func (m *KeyManager) UserKeyExists(ctx context.Context, userID string) (bool, error) {
	_, err := m.kv.Get(ctx, userKeyPath(userID))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
