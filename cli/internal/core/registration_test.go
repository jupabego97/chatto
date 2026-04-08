package core

import (
	"testing"
	"time"
)

// ============================================================================
// Registration Token Tests
// ============================================================================

func TestChattoCore_CreateRegistrationToken(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("creates token for email", func(t *testing.T) {
		token, err := core.CreateRegistrationToken(ctx, "newuser@example.com")
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
		if token == "" {
			t.Error("Expected non-empty token")
		}
		if len(token) != 16 { // "RG" prefix + 14 chars
			t.Errorf("Expected token length 16, got %d", len(token))
		}
	})

	t.Run("normalizes email to lowercase", func(t *testing.T) {
		token, err := core.CreateRegistrationToken(ctx, "  UpperCase@Example.COM  ")
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		tokenData, err := core.GetRegistrationToken(ctx, token)
		if err != nil {
			t.Fatalf("Failed to get token: %v", err)
		}
		if tokenData.Email != "uppercase@example.com" {
			t.Errorf("Expected normalized email 'uppercase@example.com', got %q", tokenData.Email)
		}
	})

	t.Run("returns error for empty email", func(t *testing.T) {
		_, err := core.CreateRegistrationToken(ctx, "")
		if err == nil {
			t.Error("Expected error for empty email")
		}
	})

	t.Run("returns error for whitespace-only email", func(t *testing.T) {
		_, err := core.CreateRegistrationToken(ctx, "   ")
		if err == nil {
			t.Error("Expected error for whitespace-only email")
		}
	})
}

func TestChattoCore_GetRegistrationToken(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("retrieves valid token", func(t *testing.T) {
		token, err := core.CreateRegistrationToken(ctx, "get-test@example.com")
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		tokenData, err := core.GetRegistrationToken(ctx, token)
		if err != nil {
			t.Fatalf("Failed to get token: %v", err)
		}
		if tokenData.Email != "get-test@example.com" {
			t.Errorf("Expected email 'get-test@example.com', got %q", tokenData.Email)
		}
		if tokenData.CreatedAt.IsZero() {
			t.Error("Expected non-zero CreatedAt")
		}
	})

	t.Run("returns error for non-existent token", func(t *testing.T) {
		_, err := core.GetRegistrationToken(ctx, "nonexistent-token")
		if err != ErrRegistrationTokenNotFound {
			t.Errorf("Expected ErrRegistrationTokenNotFound, got %v", err)
		}
	})
}

func TestChattoCore_DeleteRegistrationToken(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("deletes existing token", func(t *testing.T) {
		token, err := core.CreateRegistrationToken(ctx, "delete-test@example.com")
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Verify token exists
		_, err = core.GetRegistrationToken(ctx, token)
		if err != nil {
			t.Fatalf("Token should exist: %v", err)
		}

		// Delete token
		err = core.DeleteRegistrationToken(ctx, token)
		if err != nil {
			t.Fatalf("Failed to delete token: %v", err)
		}

		// Verify token no longer exists
		_, err = core.GetRegistrationToken(ctx, token)
		if err != ErrRegistrationTokenNotFound {
			t.Errorf("Expected ErrRegistrationTokenNotFound after delete, got %v", err)
		}
	})

	t.Run("no error when deleting non-existent token", func(t *testing.T) {
		err := core.DeleteRegistrationToken(ctx, "nonexistent-token")
		if err != nil {
			t.Errorf("Should not error when deleting non-existent token: %v", err)
		}
	})
}

func TestChattoCore_RegistrationTokenExpiration(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("fresh token is valid", func(t *testing.T) {
		token, _ := core.CreateRegistrationToken(ctx, "fresh@example.com")

		_, err := core.GetRegistrationToken(ctx, token)
		if err != nil {
			t.Fatalf("Fresh token should be valid: %v", err)
		}
	})

	t.Run("TTL is set to 24 hours", func(t *testing.T) {
		if RegistrationTokenTTL != 24*time.Hour {
			t.Errorf("Expected TTL of 24 hours, got %v", RegistrationTokenTTL)
		}
	})
}
