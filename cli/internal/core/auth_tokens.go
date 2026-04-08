package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// ============================================================================
// Auth Token Errors
// ============================================================================

var (
	// ErrAuthTokenNotFound is returned when a bearer auth token doesn't exist or has expired.
	ErrAuthTokenNotFound = errors.New("auth token not found")
)

// authTokenKeyPrefix is the KV key prefix for bearer session tokens.
const authTokenKeyPrefix = "session."

// ============================================================================
// Auth Token Types
// ============================================================================

// AuthTokenData is the JSON value stored in the AUTH_TOKENS KV bucket.
type AuthTokenData struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// Auth Token Operations
// ============================================================================

// CreateAuthToken creates a new opaque bearer token for the given user.
// The token is stored in the AUTH_TOKENS KV bucket and can be used for API authentication.
// Token expiry is handled by NATS KV TTL.
func (c *ChattoCore) CreateAuthToken(ctx context.Context, userID string) (string, error) {
	token := NewAuthToken()

	data, err := json.Marshal(AuthTokenData{
		UserID:    userID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth token: %w", err)
	}

	_, err = c.storage.authTokensKV.Put(ctx, authTokenKeyPrefix+token, data)
	if err != nil {
		return "", fmt.Errorf("failed to store auth token: %w", err)
	}

	return token, nil
}

// ValidateAuthToken checks if a bearer token is valid and returns the associated user ID.
// Returns ErrAuthTokenNotFound if the token doesn't exist (or has expired via NATS TTL).
//
// Sliding window: each successful validation re-puts the entry to reset the NATS KV TTL.
// This means the token only expires after the configured TTL of *inactivity* — active
// users are never logged out.
func (c *ChattoCore) ValidateAuthToken(ctx context.Context, token string) (string, error) {
	key := authTokenKeyPrefix + token
	entry, err := c.storage.authTokensKV.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			// MIGRATION: Try the legacy unprefixed key for tokens created before
			// the "session." prefix was added. On success, migrate the token to the
			// new key format so subsequent validations use the prefixed key.
			// TODO: Remove this fallback once all active sessions have expired or
			// after a reasonable migration window (e.g. 2 releases).
			return c.validateLegacyToken(ctx, token)
		}
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	var tokenData AuthTokenData
	if err := json.Unmarshal(entry.Value(), &tokenData); err != nil {
		return "", fmt.Errorf("failed to unmarshal auth token: %w", err)
	}

	// Re-put to reset TTL (sliding window expiry).
	// Fire-and-forget — validation succeeds even if the re-put fails.
	_, _ = c.storage.authTokensKV.Put(ctx, key, entry.Value())

	return tokenData.UserID, nil
}

// validateLegacyToken checks for a token stored under the old unprefixed key format.
// If found, it migrates the token to the new "session." prefixed key and deletes the old one.
//
// MIGRATION: Added when bearer tokens moved from raw keys to "session." prefixed keys.
// TODO: Remove after all active sessions have naturally expired (controlled by KV TTL).
func (c *ChattoCore) validateLegacyToken(ctx context.Context, token string) (string, error) {
	entry, err := c.storage.authTokensKV.Get(ctx, token)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return "", ErrAuthTokenNotFound
		}
		return "", fmt.Errorf("failed to get legacy auth token: %w", err)
	}

	var tokenData AuthTokenData
	if err := json.Unmarshal(entry.Value(), &tokenData); err != nil {
		return "", fmt.Errorf("failed to unmarshal legacy auth token: %w", err)
	}

	// Migrate: write to new prefixed key and delete the old one.
	// Both are fire-and-forget — validation succeeds even if migration fails.
	_, _ = c.storage.authTokensKV.Put(ctx, authTokenKeyPrefix+token, entry.Value())
	_ = c.storage.authTokensKV.Delete(ctx, token)

	return tokenData.UserID, nil
}

// RevokeAuthToken deletes a bearer token, immediately invalidating it.
// This is idempotent — revoking a non-existent token is not an error.
func (c *ChattoCore) RevokeAuthToken(ctx context.Context, token string) error {
	err := c.storage.authTokensKV.Delete(ctx, authTokenKeyPrefix+token)
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		return fmt.Errorf("failed to revoke auth token: %w", err)
	}
	return nil
}
