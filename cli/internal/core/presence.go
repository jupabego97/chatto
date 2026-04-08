package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// Presence status constants (matching GraphQL enum values)
const (
	PresenceStatusOffline      = "OFFLINE"
	PresenceStatusOnline       = "ONLINE"
	PresenceStatusAway         = "AWAY"
	PresenceStatusDoNotDisturb = "DO_NOT_DISTURB"
)

// Presence configuration constants
const (
	// PresenceTTL is the TTL for presence entries in the KV bucket.
	// If a client disconnects without explicit cleanup, entries expire after this duration.
	PresenceTTL = 60 * time.Second

	// PresenceRefreshInterval is how often clients refresh their presence.
	// Should be less than PresenceTTL to ensure entries don't expire while connected.
	PresenceRefreshInterval = 30 * time.Second
)

// presenceStatusFromString converts a GraphQL PresenceStatus string to protobuf enum.
// Note: OFFLINE should never be stored - callers should delete the key instead.
func presenceStatusFromString(s string) corev1.UserPresenceStatus {
	switch s {
	case PresenceStatusAway:
		return corev1.UserPresenceStatus_USER_PRESENCE_STATUS_AWAY
	case PresenceStatusDoNotDisturb:
		return corev1.UserPresenceStatus_USER_PRESENCE_STATUS_DO_NOT_DISTURB
	default:
		return corev1.UserPresenceStatus_USER_PRESENCE_STATUS_ONLINE
	}
}

// presenceStatusToString converts a protobuf UserPresenceStatus enum to GraphQL string.
func presenceStatusToString(status corev1.UserPresenceStatus) string {
	switch status {
	case corev1.UserPresenceStatus_USER_PRESENCE_STATUS_AWAY:
		return PresenceStatusAway
	case corev1.UserPresenceStatus_USER_PRESENCE_STATUS_DO_NOT_DISTURB:
		return PresenceStatusDoNotDisturb
	default:
		return PresenceStatusOnline
	}
}

// ============================================================================
// Key Helpers
// ============================================================================

// presenceKey returns the KV key for a user's presence.
func presenceKey(userID string) string {
	return fmt.Sprintf("presence.%s", userID)
}

// parseUserIDFromPresenceKey extracts the userID from a presence key.
// Key format: presence.{userId}
func parseUserIDFromPresenceKey(key string) string {
	return strings.TrimPrefix(key, "presence.")
}

// ============================================================================
// Presence Operations
// ============================================================================

// GetUserPresence retrieves a user's current presence status.
// Returns "OFFLINE" if the user has no presence entry (never connected or TTL expired).
func (c *ChattoCore) GetUserPresence(ctx context.Context, userID string) (string, error) {
	entry, err := c.storage.presenceKV.Get(ctx, presenceKey(userID))
	if err != nil {
		// Key not found means user is offline
		if err == jetstream.ErrKeyNotFound {
			return PresenceStatusOffline, nil
		}
		return PresenceStatusOffline, fmt.Errorf("failed to get presence: %w", err)
	}

	// Unmarshal protobuf payload
	presence := &corev1.UserPresence{}
	if err := proto.Unmarshal(entry.Value(), presence); err != nil {
		c.logger.Warn("Failed to unmarshal presence, treating as offline",
			"error", err, "user_id", userID)
		return PresenceStatusOffline, nil
	}

	return presenceStatusToString(presence.Status), nil
}

// SetPresence writes/refreshes a user's presence in the bucket.
// Authorization: Caller must verify the user is authenticated before calling.
func (c *ChattoCore) SetPresence(ctx context.Context, userID string, status string) error {
	// Create and marshal protobuf message
	presence := &corev1.UserPresence{
		Status: presenceStatusFromString(status),
	}

	data, err := proto.Marshal(presence)
	if err != nil {
		return fmt.Errorf("failed to marshal presence: %w", err)
	}

	_, err = c.storage.presenceKV.Put(ctx, presenceKey(userID), data)
	return err
}

// refreshPresence reads the current presence value from KV and re-puts it
// to refresh the TTL. If no entry exists (race with expiry), sets ONLINE as default.
// This preserves whatever status the client set via updateMyPresence.
//
// Uses optimistic locking (kv.Update with revision) to avoid overwriting a concurrent
// SetPresence call from updateMyPresence. If the revision has changed between Get and
// Update, the newer value is preserved and we silently skip the refresh.
func (c *ChattoCore) refreshPresence(ctx context.Context, userID string) error {
	entry, err := c.storage.presenceKV.Get(ctx, presenceKey(userID))
	if err != nil {
		if err == jetstream.ErrKeyNotFound {
			// Entry expired between ticks — re-set to ONLINE as safe default
			return c.SetPresence(ctx, userID, PresenceStatusOnline)
		}
		return fmt.Errorf("failed to read presence for refresh: %w", err)
	}

	// Re-put the same value to refresh TTL using optimistic locking.
	// If a concurrent SetPresence modified the entry, Update fails and
	// the newer status is preserved — which is the correct behavior.
	_, err = c.storage.presenceKV.Update(ctx, presenceKey(userID), entry.Value(), entry.Revision())
	if err != nil {
		// ErrKeyExists means the revision changed (concurrent write) — that's fine,
		// the newer value already has a fresh TTL from the concurrent Put.
		if err == jetstream.ErrKeyExists {
			return nil
		}
		return fmt.Errorf("failed to refresh presence: %w", err)
	}
	return nil
}

// kvEntryToPresenceChange converts a KV entry to a PresenceChange proto.
// Handles both PUT operations (user came online/changed status) and DELETE operations (user went offline).
func (c *ChattoCore) kvEntryToPresenceChange(entry jetstream.KeyValueEntry) *corev1.PresenceChange {
	// Extract userID from key (format: presence.{userId})
	userID := parseUserIDFromPresenceKey(entry.Key())

	// Handle deletion (TTL expiry or explicit delete)
	if entry.Operation() == jetstream.KeyValueDelete ||
		entry.Operation() == jetstream.KeyValuePurge {
		return &corev1.PresenceChange{
			UserId: userID,
			Status: PresenceStatusOffline,
		}
	}

	// Unmarshal protobuf payload
	presence := &corev1.UserPresence{}
	if err := proto.Unmarshal(entry.Value(), presence); err != nil {
		c.logger.Warn("Failed to unmarshal presence change, treating as offline",
			"error", err, "user_id", userID)
		return &corev1.PresenceChange{
			UserId: userID,
			Status: PresenceStatusOffline,
		}
	}

	return &corev1.PresenceChange{
		UserId: userID,
		Status: presenceStatusToString(presence.Status),
	}
}

