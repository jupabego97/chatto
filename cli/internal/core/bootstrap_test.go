package core

import (
	"errors"
	"sync"
	"testing"
)

// ============================================================================
// IsInstanceFresh Tests
// ============================================================================

func TestChattoCore_IsInstanceFresh(t *testing.T) {
	t.Run("returns true for fresh instance", func(t *testing.T) {
		core, _ := setupTestCore(t)
		ctx := testContext(t)

		fresh, err := core.IsInstanceFresh(ctx)
		if err != nil {
			t.Fatalf("Failed to check IsInstanceFresh: %v", err)
		}
		if !fresh {
			t.Error("Expected fresh instance before any bootstrap")
		}
	})

	t.Run("returns false after bootstrap", func(t *testing.T) {
		core, _ := setupTestCore(t)
		ctx := testContext(t)

		// Bootstrap the instance
		// Note: Using "bootstrapper" instead of "admin" because "admin" is blocked by default
		_, err := core.Bootstrap(ctx, BootstrapInput{
			Login:    "bootstrapper",
			Email:    "bootstrapper@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Failed to bootstrap: %v", err)
		}

		fresh, err := core.IsInstanceFresh(ctx)
		if err != nil {
			t.Fatalf("Failed to check IsInstanceFresh: %v", err)
		}
		if fresh {
			t.Error("Expected instance to NOT be fresh after bootstrap")
		}
	})
}

// ============================================================================
// Bootstrap Tests
// ============================================================================

func TestChattoCore_Bootstrap_Success(t *testing.T) {
	t.Run("creates admin user with defaults", func(t *testing.T) {
		core, _ := setupTestCore(t)
		ctx := testContext(t)

		// Note: Using "bootstrapper" instead of "admin" because "admin" is blocked by default
		result, err := core.Bootstrap(ctx, BootstrapInput{
			Login:    "bootstrapper",
			Email:    "bootstrapper@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Bootstrap failed: %v", err)
		}

		// Check user was created
		if result.User == nil {
			t.Fatal("Expected user to be created")
		}
		if result.User.Login != "bootstrapper" {
			t.Errorf("Expected login 'bootstrapper', got '%s'", result.User.Login)
		}
		if result.User.DisplayName != "bootstrapper" {
			t.Errorf("Expected display name to default to login 'bootstrapper', got '%s'", result.User.DisplayName)
		}

		// Check no space was created (not requested)
		if result.Space != nil {
			t.Error("Expected no space to be created when spaceName not provided")
		}

		// Verify user is owner (bootstrapped users become instance-owner)
		isOwner, err := core.IsInstanceOwner(ctx, result.User.Id)
		if err != nil {
			t.Fatalf("Failed to check owner status: %v", err)
		}
		if !isOwner {
			t.Error("Expected bootstrapped user to be owner")
		}

		// Verify email was added as verified
		hasVerified, err := core.HasVerifiedEmail(ctx, result.User.Id)
		if err != nil {
			t.Fatalf("Failed to check verified status: %v", err)
		}
		if !hasVerified {
			t.Error("Expected bootstrapped user to have verified email")
		}
	})

	t.Run("creates admin user with custom display name", func(t *testing.T) {
		core, _ := setupTestCore(t)
		ctx := testContext(t)

		// Note: Using "bootstrapper" instead of "admin" because "admin" is blocked by default
		result, err := core.Bootstrap(ctx, BootstrapInput{
			Login:       "bootstrapper",
			DisplayName: "Admin User",
			Email:       "bootstrapper@example.com",
			Password:    "password123",
		})
		if err != nil {
			t.Fatalf("Bootstrap failed: %v", err)
		}

		if result.User.DisplayName != "Admin User" {
			t.Errorf("Expected display name 'Admin User', got '%s'", result.User.DisplayName)
		}
	})

	t.Run("creates initial space with default rooms", func(t *testing.T) {
		core, _ := setupTestCore(t)
		ctx := testContext(t)

		// Note: Using "bootstrapper" instead of "admin" because "admin" is blocked by default
		result, err := core.Bootstrap(ctx, BootstrapInput{
			Login:            "bootstrapper",
			Email:            "bootstrapper@example.com",
			Password:         "password123",
			SpaceName:        "My Community",
			SpaceDescription: "A great place to hang out",
		})
		if err != nil {
			t.Fatalf("Bootstrap failed: %v", err)
		}

		// Check space was created
		if result.Space == nil {
			t.Fatal("Expected space to be created")
		}
		if result.Space.Name != "My Community" {
			t.Errorf("Expected space name 'My Community', got '%s'", result.Space.Name)
		}
		if result.Space.Description != "A great place to hang out" {
			t.Errorf("Expected space description 'A great place to hang out', got '%s'", result.Space.Description)
		}

		// Check default rooms were created
		rooms, err := core.ListRoomsBySpace(ctx, result.Space.Id)
		if err != nil {
			t.Fatalf("Failed to list rooms: %v", err)
		}

		roomNames := make(map[string]bool)
		for _, room := range rooms {
			roomNames[room.Name] = true
		}

		// Should have the default auto-join rooms
		for _, expectedRoom := range DefaultAutoJoinRoomNames {
			if !roomNames[expectedRoom] {
				t.Errorf("Expected default room '%s' to be created", expectedRoom)
			}
		}

		// Verify bootstrap-created rooms have auto_join enabled
		for _, room := range rooms {
			if roomNames[room.Name] {
				if !room.AutoJoin {
					t.Errorf("Expected bootstrap room '%s' to have auto_join=true", room.Name)
				}
			}
		}
	})

	t.Run("user can authenticate after bootstrap", func(t *testing.T) {
		core, _ := setupTestCore(t)
		ctx := testContext(t)

		// Note: Using "bootstrapper" instead of "admin" because "admin" is blocked by default
		_, err := core.Bootstrap(ctx, BootstrapInput{
			Login:    "bootstrapper",
			Email:    "bootstrapper@example.com",
			Password: "secretpassword",
		})
		if err != nil {
			t.Fatalf("Bootstrap failed: %v", err)
		}

		// Try to authenticate
		user, err := core.VerifyPassword(ctx, "bootstrapper", "secretpassword")
		if err != nil {
			t.Fatalf("Authentication failed: %v", err)
		}
		if user == nil {
			t.Error("Expected to authenticate successfully")
		}
		if user != nil && user.Login != "bootstrapper" {
			t.Errorf("Expected authenticated user to be 'bootstrapper', got '%s'", user.Login)
		}
	})
}

func TestChattoCore_Bootstrap_AlreadyBootstrapped(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// First bootstrap should succeed
	// Note: Using "bootstrapper" instead of "admin" because "admin" is blocked by default
	_, err := core.Bootstrap(ctx, BootstrapInput{
		Login:    "bootstrapper",
		Email:    "bootstrapper@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("First bootstrap failed: %v", err)
	}

	// Second bootstrap should fail
	_, err = core.Bootstrap(ctx, BootstrapInput{
		Login:    "bootstrapper2",
		Email:    "bootstrapper2@example.com",
		Password: "password456",
	})
	if err == nil {
		t.Fatal("Expected second bootstrap to fail")
	}
	if !errors.Is(err, ErrAlreadyBootstrapped) {
		t.Errorf("Expected ErrAlreadyBootstrapped, got: %v", err)
	}
}

func TestChattoCore_Bootstrap_Concurrent(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	const numGoroutines = 10
	results := make(chan struct {
		user string
		err  error
	}, numGoroutines)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Race multiple goroutines trying to bootstrap
	// Note: Using "boot" prefix instead of "admin" because "admin" is blocked by default
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			result, err := core.Bootstrap(ctx, BootstrapInput{
				Login:    "boot" + string(rune('a'+idx)),
				Email:    "boot" + string(rune('a'+idx)) + "@example.com",
				Password: "password123",
			})
			userID := ""
			if result != nil && result.User != nil {
				userID = result.User.Id
			}
			results <- struct {
				user string
				err  error
			}{userID, err}
		}(i)
	}

	wg.Wait()
	close(results)

	// Collect results
	successCount := 0
	alreadyBootstrappedCount := 0
	var successUserID string

	for r := range results {
		if r.err == nil {
			successCount++
			successUserID = r.user
		} else if errors.Is(r.err, ErrAlreadyBootstrapped) {
			alreadyBootstrappedCount++
		} else {
			t.Errorf("Unexpected error: %v", r.err)
		}
	}

	// Exactly one should succeed
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful bootstrap, got %d", successCount)
	}

	// The rest should get ErrAlreadyBootstrapped
	if alreadyBootstrappedCount != numGoroutines-1 {
		t.Errorf("Expected %d ErrAlreadyBootstrapped, got %d", numGoroutines-1, alreadyBootstrappedCount)
	}

	// Verify the successful user is actually owner
	if successUserID != "" {
		isOwner, _ := core.IsInstanceOwner(ctx, successUserID)
		if !isOwner {
			t.Error("Expected successful bootstrap user to be owner")
		}
	}
}
