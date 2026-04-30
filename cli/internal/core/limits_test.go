package core

import (
	"errors"
	"testing"

	"hmans.de/chatto/internal/config"
)

func TestCreateSpace_RespectsMaxSpacesLimit(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	user, err := core.CreateUser(ctx, "system", "limit-user", "Limit User", "password123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	zero := 0
	core.config.Limits = config.LimitsConfig{MaxSpaces: &zero}
	if _, err := core.CreateSpace(ctx, user.Id, "Locked", ""); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded with max=0, got %v", err)
	}

	two := 2
	core.config.Limits = config.LimitsConfig{MaxSpaces: &two}

	if _, err := core.CreateSpace(ctx, user.Id, "First", ""); err != nil {
		t.Fatalf("first space should succeed: %v", err)
	}
	if _, err := core.CreateSpace(ctx, user.Id, "Second", ""); err != nil {
		t.Fatalf("second space should succeed: %v", err)
	}
	if _, err := core.CreateSpace(ctx, user.Id, "Third", ""); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("third space should be blocked, got %v", err)
	}
}

func TestCreateSpace_UnlimitedByDefault(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	user, err := core.CreateUser(ctx, "system", "unlim-user", "Unlim User", "password123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	for i, name := range []string{"a", "b", "c", "d"} {
		if _, err := core.CreateSpace(ctx, user.Id, name, ""); err != nil {
			t.Fatalf("space %d (%q) should succeed under default unlimited: %v", i, name, err)
		}
	}
}

func TestCreateUser_RespectsMaxUsersLimit(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Create + verify the first user before applying the limit.
	u1, err := core.CreateUser(ctx, "system", "signup-user-1", "U1", "password123")
	if err != nil {
		t.Fatalf("create u1: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, u1.Id, "u1@example.com"); err != nil {
		t.Fatalf("verify u1: %v", err)
	}

	// Now lock the door at 1 verified user.
	one := 1
	core.config.Limits = config.LimitsConfig{MaxUsers: &one}

	if _, err := core.CreateUser(ctx, "system", "signup-user-2", "U2", "password123"); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("signup should be blocked when at verified-user limit, got %v", err)
	}
}

func TestAddVerifiedEmail_AtomicLimitGate(t *testing.T) {
	// The verification-time gate uses CAS-incrementing the stats counter, so it
	// catches concurrent signups that both passed the (non-atomic) signup gate.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Two users created with the limit disabled, then we lock at 1 and try to
	// verify both. Only the first verification should succeed.
	u1, _ := core.CreateUser(ctx, "system", "race-u1", "U1", "password123")
	u2, _ := core.CreateUser(ctx, "system", "race-u2", "U2", "password123")

	one := 1
	core.config.Limits = config.LimitsConfig{MaxUsers: &one}

	if err := core.AddVerifiedEmailDirect(ctx, u1.Id, "u1@example.com"); err != nil {
		t.Fatalf("first verification should succeed: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, u2.Id, "u2@example.com"); !errors.Is(err, ErrLimitExceeded) {
		t.Errorf("second verification should hit the atomic gate, got %v", err)
	}

	// The verified_users counter must reflect exactly 1 — no overshoot.
	got, _ := core.GetStat(ctx, StatVerifiedUsers)
	if got != 1 {
		t.Errorf("verified_users stat = %d, want 1", got)
	}
}

func TestAddVerifiedEmail_AdditionalEmailDoesNotIncrement(t *testing.T) {
	// Adding a second verified email to an already-verified user must not
	// re-trip the limit gate or double-count in the stats counter.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	one := 1
	core.config.Limits = config.LimitsConfig{MaxUsers: &one}

	u, _ := core.CreateUser(ctx, "system", "multi-email-user", "Multi", "password123")
	if err := core.AddVerifiedEmailDirect(ctx, u.Id, "primary@example.com"); err != nil {
		t.Fatalf("first email should verify: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, u.Id, "secondary@example.com"); err != nil {
		t.Fatalf("second email on already-verified user should not be blocked: %v", err)
	}

	got, _ := core.GetStat(ctx, StatVerifiedUsers)
	if got != 1 {
		t.Errorf("verified_users stat = %d, want 1 (transition only counts once)", got)
	}
}

func TestDeleteUser_DecrementsVerifiedUsers(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	u, _ := core.CreateUser(ctx, "system", "delete-test-user", "D", "password123")
	if err := core.AddVerifiedEmailDirect(ctx, u.Id, "delete@example.com"); err != nil {
		t.Fatalf("verify: %v", err)
	}

	before, _ := core.GetStat(ctx, StatVerifiedUsers)
	if before != 1 {
		t.Fatalf("setup: stat = %d, want 1", before)
	}

	if err := core.DeleteUser(ctx, "system", u.Id); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	after, _ := core.GetStat(ctx, StatVerifiedUsers)
	if after != 0 {
		t.Errorf("verified_users stat after delete = %d, want 0", after)
	}
}

func TestDeleteSpace_DecrementsSpaces(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	u, _ := core.CreateUser(ctx, "system", "del-space-user", "DS", "password123")
	sp, err := core.CreateSpace(ctx, u.Id, "DeleteMe", "")
	if err != nil {
		t.Fatalf("create space: %v", err)
	}

	before, _ := core.GetStat(ctx, StatSpaces)
	if before != 1 {
		t.Fatalf("setup: spaces stat = %d, want 1", before)
	}

	if err := core.DeleteSpace(ctx, u.Id, sp.Id); err != nil {
		t.Fatalf("delete space: %v", err)
	}

	after, _ := core.GetStat(ctx, StatSpaces)
	if after != 0 {
		t.Errorf("spaces stat after delete = %d, want 0", after)
	}
}

func TestCountSpacesAndUsers(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// CountSpaces excludes the DM system space and CountVerifiedUsers starts at 0.
	if got, _ := core.CountSpaces(ctx); got != 0 {
		t.Errorf("baseline CountSpaces = %d, want 0", got)
	}
	if got, _ := core.CountVerifiedUsers(ctx); got != 0 {
		t.Errorf("baseline CountVerifiedUsers = %d, want 0", got)
	}

	u, _ := core.CreateUser(ctx, "system", "count-user", "Count", "password123")
	if _, err := core.CreateSpace(ctx, u.Id, "S1", ""); err != nil {
		t.Fatalf("create space: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, u.Id, "count@example.com"); err != nil {
		t.Fatalf("verify email: %v", err)
	}

	if got, _ := core.CountSpaces(ctx); got != 1 {
		t.Errorf("CountSpaces = %d, want 1", got)
	}
	if got, _ := core.CountVerifiedUsers(ctx); got != 1 {
		t.Errorf("CountVerifiedUsers = %d, want 1", got)
	}
}
