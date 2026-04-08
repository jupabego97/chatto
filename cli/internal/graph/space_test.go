package graph

import (
	"errors"
	"testing"
)

// ============================================================================
// Space Field Resolver Tests
// ============================================================================

func TestSpaceResolver_Rooms(t *testing.T) {
	env := setupTestResolver(t)

	t.Run("list rooms for space (authorized)", func(t *testing.T) {
		rooms, err := env.resolver.Space().Rooms(env.authContext(), env.testSpace)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(rooms) == 0 {
			t.Fatal("Expected at least one room")
		}

		// Verify test room is in the list
		found := false
		for _, room := range rooms {
			if room.Id == env.testRoom.Id {
				found = true
				break
			}
		}

		if !found {
			t.Error("Test room not found in rooms list")
		}
	})

	t.Run("list rooms for space (unauthorized - not a member)", func(t *testing.T) {
		// Create a user who is not a member
		user2, err := env.core.CreateUser(env.ctx, "system", "outsider", "outsider", "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		rooms, err := env.resolver.Space().Rooms(env.authContextForUser(user2), env.testSpace)
		if !errors.Is(err, ErrNotSpaceMember) {
			t.Errorf("Expected ErrNotSpaceMember, got %v", err)
		}

		if rooms != nil {
			t.Errorf("Expected nil rooms, got %+v", rooms)
		}
	})
}

// ============================================================================
// Room Field Resolver Tests
// ============================================================================

func TestRoomResolver_Members(t *testing.T) {
	env := setupTestResolver(t)

	t.Run("room member can list members", func(t *testing.T) {
		members, err := env.resolver.Room().Members(env.authContext(), env.testRoom)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if members == nil {
			t.Fatal("Expected members, got nil")
		}
		// Should have at least the test user
		if len(members) == 0 {
			t.Error("Expected at least one member")
		}
	})

	t.Run("unauthenticated user is rejected", func(t *testing.T) {
		members, err := env.resolver.Room().Members(env.unauthContext(), env.testRoom)
		if !errors.Is(err, ErrNotAuthenticated) {
			t.Errorf("Expected ErrNotAuthenticated, got %v", err)
		}
		if members != nil {
			t.Errorf("Expected nil members, got %+v", members)
		}
	})

	t.Run("non-room-member is rejected", func(t *testing.T) {
		outsider, err := env.core.CreateUser(env.ctx, "system", "outsider-members", "Outsider", "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		members, err := env.resolver.Room().Members(env.authContextForUser(outsider), env.testRoom)
		if !errors.Is(err, ErrNotRoomMember) {
			t.Errorf("Expected ErrNotRoomMember, got %v", err)
		}
		if members != nil {
			t.Errorf("Expected nil members, got %+v", members)
		}
	})

	t.Run("space member but not room member is rejected", func(t *testing.T) {
		spaceMember, err := env.core.CreateUser(env.ctx, "system", "spacemember-members", "Space Member", "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		_, err = env.core.JoinSpace(env.ctx, spaceMember.Id, env.testSpace.Id)
		if err != nil {
			t.Fatalf("Failed to join space: %v", err)
		}

		members, err := env.resolver.Room().Members(env.authContextForUser(spaceMember), env.testRoom)
		if !errors.Is(err, ErrNotRoomMember) {
			t.Errorf("Expected ErrNotRoomMember, got %v", err)
		}
		if members != nil {
			t.Errorf("Expected nil members, got %+v", members)
		}
	})
}
