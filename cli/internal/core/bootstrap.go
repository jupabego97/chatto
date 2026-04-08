package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// BootstrapInput contains all data needed to bootstrap a fresh instance.
type BootstrapInput struct {
	Login            string
	DisplayName      string // Optional - defaults to Login if empty
	Email            string
	Password         string
	SpaceName        string // Optional
	SpaceDescription string // Optional
}

// BootstrapResult contains the created entities.
type BootstrapResult struct {
	User  *corev1.User
	Space *corev1.Space // May be nil if no space was requested
}

// ErrAlreadyBootstrapped is returned when bootstrap is called on a non-fresh instance.
var ErrAlreadyBootstrapped = fmt.Errorf("instance has already been bootstrapped")

// Bootstrap atomically initializes a fresh Chatto instance.
// Creates the first admin user and optionally an initial space with default rooms.
// Returns ErrAlreadyBootstrapped if instance was already set up.
//
// Authorization: None required (this creates the first user).
func (c *ChattoCore) Bootstrap(ctx context.Context, input BootstrapInput) (*BootstrapResult, error) {
	kv := c.storage.instanceRBACKV

	// Step 1: Atomically claim first owner marker (race condition protection)
	// We'll use a temporary marker to claim the bootstrap operation
	// The actual user ID will replace this after user creation succeeds
	_, err := kv.Create(ctx, firstOwnerMarkerKey, []byte("bootstrapping"))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyExists) {
			return nil, ErrAlreadyBootstrapped
		}
		return nil, fmt.Errorf("failed to claim bootstrap: %w", err)
	}

	// Step 2: Create the owner user
	displayName := input.DisplayName
	if displayName == "" {
		displayName = input.Login
	}
	user, err := c.CreateUser(ctx, "system", input.Login, displayName, input.Password)
	if err != nil {
		// Rollback: delete the marker so another attempt can be made
		_ = kv.Delete(ctx, firstOwnerMarkerKey)
		return nil, fmt.Errorf("failed to create owner user: %w", err)
	}

	// Step 3: Update marker with actual user ID
	_, err = kv.Put(ctx, firstOwnerMarkerKey, []byte(user.Id))
	if err != nil {
		c.logger.Warn("Failed to update bootstrap marker with user ID", "error", err)
		// Continue anyway - marker exists, user created
	}

	// Step 4: Assign owner role to user
	if err := c.AssignInstanceOwnerRole(ctx, user.Id); err != nil {
		c.logger.Warn("Failed to assign owner role during bootstrap", "error", err)
		// Continue anyway - user created, can fix role later
	}

	// Step 5: Add verified email directly (no verification needed for bootstrap)
	if input.Email != "" {
		if err := c.AddVerifiedEmailDirect(ctx, user.Id, input.Email); err != nil {
			c.logger.Warn("Failed to add verified email during bootstrap", "error", err)
			// Continue anyway - user created
		}
	}

	result := &BootstrapResult{User: user}

	// Step 6: Optionally create initial space
	if input.SpaceName != "" {
		space, err := c.CreateSpace(ctx, user.Id, input.SpaceName, input.SpaceDescription)
		if err != nil {
			c.logger.Warn("Failed to create initial space during bootstrap", "error", err)
			// Continue anyway - user created
		} else {
			// Create default rooms in the space and join the user to them.
			// Note: CreateSpace calls autoJoinDefaultRooms, but at that point the rooms
			// don't exist yet, so we need to join the user after creating the rooms.
			for _, roomName := range DefaultAutoJoinRoomNames {
				room, err := c.CreateRoom(ctx, user.Id, space.Id, roomName, "")
				if err != nil {
					c.logger.Warn("Failed to create default room during bootstrap",
						"room", roomName, "error", err)
					continue
				}
				// Mark room as auto-join so new members automatically join it
				if _, err := c.SetRoomAutoJoin(ctx, user.Id, space.Id, room.Id, true); err != nil {
					c.logger.Warn("Failed to set auto_join on default room during bootstrap",
						"room", roomName, "error", err)
				}
				// Join the admin user to this room
				_, err = c.JoinRoom(ctx, user.Id, space.Id, user.Id, room.Id)
				if err != nil {
					c.logger.Warn("Failed to join admin to default room during bootstrap",
						"room", roomName, "error", err)
				}
			}
			result.Space = space
		}
	}

	c.logger.Info("Instance bootstrapped successfully", "admin_id", user.Id)
	return result, nil
}
