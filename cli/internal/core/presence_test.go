package core

import (
	"testing"

	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// ============================================================================
// Status Conversion Tests
// ============================================================================

func TestPresenceStatusFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected corev1.UserPresenceStatus
	}{
		{
			name:     "ONLINE status",
			input:    PresenceStatusOnline,
			expected: corev1.UserPresenceStatus_USER_PRESENCE_STATUS_ONLINE,
		},
		{
			name:     "AWAY status",
			input:    PresenceStatusAway,
			expected: corev1.UserPresenceStatus_USER_PRESENCE_STATUS_AWAY,
		},
		{
			name:     "DO_NOT_DISTURB status",
			input:    PresenceStatusDoNotDisturb,
			expected: corev1.UserPresenceStatus_USER_PRESENCE_STATUS_DO_NOT_DISTURB,
		},
		{
			name:     "unknown status defaults to ONLINE",
			input:    "UNKNOWN",
			expected: corev1.UserPresenceStatus_USER_PRESENCE_STATUS_ONLINE,
		},
		{
			name:     "empty string defaults to ONLINE",
			input:    "",
			expected: corev1.UserPresenceStatus_USER_PRESENCE_STATUS_ONLINE,
		},
		{
			name:     "OFFLINE defaults to ONLINE (should not be stored)",
			input:    PresenceStatusOffline,
			expected: corev1.UserPresenceStatus_USER_PRESENCE_STATUS_ONLINE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := presenceStatusFromString(tt.input)
			if result != tt.expected {
				t.Errorf("presenceStatusFromString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPresenceStatusToString(t *testing.T) {
	tests := []struct {
		name     string
		input    corev1.UserPresenceStatus
		expected string
	}{
		{
			name:     "ONLINE status",
			input:    corev1.UserPresenceStatus_USER_PRESENCE_STATUS_ONLINE,
			expected: PresenceStatusOnline,
		},
		{
			name:     "AWAY status",
			input:    corev1.UserPresenceStatus_USER_PRESENCE_STATUS_AWAY,
			expected: PresenceStatusAway,
		},
		{
			name:     "DO_NOT_DISTURB status",
			input:    corev1.UserPresenceStatus_USER_PRESENCE_STATUS_DO_NOT_DISTURB,
			expected: PresenceStatusDoNotDisturb,
		},
		{
			name:     "unknown enum value defaults to ONLINE",
			input:    corev1.UserPresenceStatus(999),
			expected: PresenceStatusOnline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := presenceStatusToString(tt.input)
			if result != tt.expected {
				t.Errorf("presenceStatusToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPresenceStatusRoundTrip(t *testing.T) {
	// Verify that converting to proto and back yields the same string
	statuses := []string{PresenceStatusOnline, PresenceStatusAway, PresenceStatusDoNotDisturb}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			proto := presenceStatusFromString(status)
			result := presenceStatusToString(proto)
			if result != status {
				t.Errorf("Round trip failed: %q -> %v -> %q", status, proto, result)
			}
		})
	}
}

// ============================================================================
// Key Helper Tests
// ============================================================================

func TestPresenceKey(t *testing.T) {
	tests := []struct {
		userID   string
		expected string
	}{
		{"user123", "presence.user123"},
		{"abc", "presence.abc"},
		{"a1b2c3d4e5f6g7", "presence.a1b2c3d4e5f6g7"},
	}

	for _, tt := range tests {
		t.Run(tt.userID, func(t *testing.T) {
			result := presenceKey(tt.userID)
			if result != tt.expected {
				t.Errorf("presenceKey(%q) = %q, want %q", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestParseUserIDFromPresenceKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"presence.user123", "user123"},
		{"presence.abc", "abc"},
		{"presence.a1b2c3d4e5f6g7", "a1b2c3d4e5f6g7"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := parseUserIDFromPresenceKey(tt.key)
			if result != tt.expected {
				t.Errorf("parseUserIDFromPresenceKey(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestPresenceKeyRoundTrip(t *testing.T) {
	userIDs := []string{"user123", "abc", "a1b2c3d4e5f6g7"}

	for _, userID := range userIDs {
		t.Run(userID, func(t *testing.T) {
			key := presenceKey(userID)
			result := parseUserIDFromPresenceKey(key)
			if result != userID {
				t.Errorf("Round trip failed: %q -> %q -> %q", userID, key, result)
			}
		})
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestChattoCore_GetUserPresence_Offline(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// User with no presence entry should be OFFLINE
	status, err := core.GetUserPresence(ctx, "nonexistent-user")
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusOffline {
		t.Errorf("Expected OFFLINE for non-existent user, got %q", status)
	}
}

func TestChattoCore_SetAndGetPresence(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	userID := "test-user-123"

	// Set presence to ONLINE
	err := core.SetPresence(ctx, userID, PresenceStatusOnline)
	if err != nil {
		t.Fatalf("setPresence failed: %v", err)
	}

	// Verify presence is ONLINE
	status, err := core.GetUserPresence(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusOnline {
		t.Errorf("Expected ONLINE, got %q", status)
	}

	// Change to AWAY
	err = core.SetPresence(ctx, userID, PresenceStatusAway)
	if err != nil {
		t.Fatalf("setPresence failed: %v", err)
	}

	status, err = core.GetUserPresence(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusAway {
		t.Errorf("Expected AWAY, got %q", status)
	}

	// Change to DO_NOT_DISTURB
	err = core.SetPresence(ctx, userID, PresenceStatusDoNotDisturb)
	if err != nil {
		t.Fatalf("setPresence failed: %v", err)
	}

	status, err = core.GetUserPresence(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusDoNotDisturb {
		t.Errorf("Expected DO_NOT_DISTURB, got %q", status)
	}
}

func TestChattoCore_PresenceDelete(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	userID := "test-user-delete"

	// Set presence
	err := core.SetPresence(ctx, userID, PresenceStatusOnline)
	if err != nil {
		t.Fatalf("setPresence failed: %v", err)
	}

	// Verify it's set
	status, _ := core.GetUserPresence(ctx, userID)
	if status != PresenceStatusOnline {
		t.Fatalf("Expected ONLINE before delete, got %q", status)
	}

	// Delete the presence entry
	err = core.storage.presenceKV.Delete(ctx, presenceKey(userID))
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should be OFFLINE now
	status, err = core.GetUserPresence(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusOffline {
		t.Errorf("Expected OFFLINE after delete, got %q", status)
	}
}

func TestChattoCore_RefreshPresence(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	userID := "test-user-refresh"

	// Set presence to AWAY
	err := core.SetPresence(ctx, userID, PresenceStatusAway)
	if err != nil {
		t.Fatalf("SetPresence failed: %v", err)
	}

	// Refresh should preserve the AWAY status
	err = core.refreshPresence(ctx, userID)
	if err != nil {
		t.Fatalf("refreshPresence failed: %v", err)
	}

	status, err := core.GetUserPresence(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusAway {
		t.Errorf("Expected AWAY after refresh, got %q", status)
	}
}

func TestChattoCore_RefreshPresence_Expired(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	userID := "test-user-refresh-expired"

	// Don't set any presence — key doesn't exist
	// refreshPresence should fall back to ONLINE
	err := core.refreshPresence(ctx, userID)
	if err != nil {
		t.Fatalf("refreshPresence failed: %v", err)
	}

	status, err := core.GetUserPresence(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserPresence failed: %v", err)
	}
	if status != PresenceStatusOnline {
		t.Errorf("Expected ONLINE as fallback, got %q", status)
	}
}

func TestChattoCore_MultipleUsersPresence(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Create multiple users
	users := make([]string, 5)
	for i := 0; i < 5; i++ {
		user, err := core.CreateUser(ctx, "system",
			"multiuser"+string(rune('0'+i)),
			"Multi User "+string(rune('0'+i)),
			"password123")
		if err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
		users[i] = user.Id
	}

	// Set different presence statuses
	statuses := []string{
		PresenceStatusOnline,
		PresenceStatusAway,
		PresenceStatusDoNotDisturb,
		PresenceStatusOnline,
		PresenceStatusAway,
	}

	for i, userID := range users {
		err := core.SetPresence(ctx, userID, statuses[i])
		if err != nil {
			t.Fatalf("Failed to set presence for user %d: %v", i, err)
		}
	}

	// Verify all statuses are correct
	for i, userID := range users {
		status, err := core.GetUserPresence(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get presence for user %d: %v", i, err)
		}
		if status != statuses[i] {
			t.Errorf("User %d: expected %q, got %q", i, statuses[i], status)
		}
	}
}
