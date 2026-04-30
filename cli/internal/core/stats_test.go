package core

import (
	"errors"
	"sync"
	"testing"
)

func TestStats_GetIsZeroWhenMissing(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	got, err := core.GetStat(ctx, "never-set")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != 0 {
		t.Errorf("missing stat = %d, want 0", got)
	}
}

func TestStats_IncrementAndDecrement(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := int64(1); i <= 3; i++ {
		got, err := core.IncrementStat(ctx, "counter")
		if err != nil {
			t.Fatalf("increment %d: %v", i, err)
		}
		if got != i {
			t.Errorf("increment returned %d, want %d", got, i)
		}
	}

	got, err := core.DecrementStat(ctx, "counter")
	if err != nil {
		t.Fatalf("decrement: %v", err)
	}
	if got != 2 {
		t.Errorf("decrement returned %d, want 2", got)
	}

	stored, _ := core.GetStat(ctx, "counter")
	if stored != 2 {
		t.Errorf("stored value = %d, want 2", stored)
	}
}

func TestStats_DecrementFlooredAtZero(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	got, err := core.DecrementStat(ctx, "missing-counter")
	if err != nil {
		t.Fatalf("decrement on missing: %v", err)
	}
	if got != 0 {
		t.Errorf("decrement on missing = %d, want 0", got)
	}

	core.IncrementStat(ctx, "floor")
	core.DecrementStat(ctx, "floor")
	got, _ = core.DecrementStat(ctx, "floor")
	if got != 0 {
		t.Errorf("over-decrement = %d, want 0 (floored)", got)
	}
}

func TestStats_IncrementStatIfBelow_AllowsThenDenies(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := int64(1); i <= 3; i++ {
		got, err := core.IncrementStatIfBelow(ctx, "gated", 3)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got != i {
			t.Errorf("call %d returned %d, want %d", i, got, i)
		}
	}

	_, err := core.IncrementStatIfBelow(ctx, "gated", 3)
	if !errors.Is(err, ErrLimitExceeded) {
		t.Errorf("4th call: expected ErrLimitExceeded, got %v", err)
	}
}

func TestStats_IncrementStatIfBelow_LockedAtZero(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	_, err := core.IncrementStatIfBelow(ctx, "locked", 0)
	if !errors.Is(err, ErrLimitExceeded) {
		t.Errorf("expected ErrLimitExceeded with max=0, got %v", err)
	}
}

func TestStats_IncrementStatIfBelow_NegativeMaxDisablesGate(t *testing.T) {
	// max < 0 is the "no limit" sentinel — useful for callers that have an
	// optional limit and want to use the same code path either way.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := int64(1); i <= 5; i++ {
		got, err := core.IncrementStatIfBelow(ctx, "unlimited", -1)
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if got != i {
			t.Errorf("call %d = %d, want %d", i, got, i)
		}
	}
}

func TestStats_ConcurrentIncrementIsAtomic(t *testing.T) {
	// Hammer the same counter from many goroutines and verify the final value
	// equals the number of increments — i.e. CAS retries don't lose any.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	const goroutines = 50

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := core.IncrementStat(ctx, "race-counter"); err != nil {
				t.Errorf("increment: %v", err)
			}
		}()
	}
	wg.Wait()

	got, _ := core.GetStat(ctx, "race-counter")
	if got != int64(goroutines) {
		t.Errorf("after %d concurrent increments: got %d", goroutines, got)
	}
}

func TestStats_ConcurrentLimitGateAllowsExactlyMax(t *testing.T) {
	// Race many goroutines through IncrementStatIfBelow and verify exactly max
	// see allowed increments — no over-shoot, no under-shoot.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	const max = int64(10)
	const goroutines = 50

	var wg sync.WaitGroup
	allowedCount := int64(0)
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := core.IncrementStatIfBelow(ctx, "race-gate", max); err == nil {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowedCount != max {
		t.Errorf("expected exactly %d allowed under contention, got %d", max, allowedCount)
	}
	stored, _ := core.GetStat(ctx, "race-gate")
	if stored != max {
		t.Errorf("stored = %d, want %d", stored, max)
	}
}

func TestStats_RecomputeFromAuthoritativeState(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Create some real state.
	user, _ := core.CreateUser(ctx, "system", "stat-user", "Stat", "password123")
	if _, err := core.CreateSpace(ctx, user.Id, "S1", ""); err != nil {
		t.Fatalf("create space: %v", err)
	}
	if _, err := core.CreateSpace(ctx, user.Id, "S2", ""); err != nil {
		t.Fatalf("create space: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, user.Id, "stat@example.com"); err != nil {
		t.Fatalf("verify: %v", err)
	}

	// Pretend the counters are wrong.
	core.setStat(ctx, StatSpaces, 999)
	core.setStat(ctx, StatVerifiedUsers, 999)

	if err := core.RecomputeStats(ctx); err != nil {
		t.Fatalf("recompute: %v", err)
	}

	gotSpaces, _ := core.GetStat(ctx, StatSpaces)
	gotUsers, _ := core.GetStat(ctx, StatVerifiedUsers)

	expectedSpaces, _ := core.CountSpaces(ctx)
	expectedUsers, _ := core.CountVerifiedUsers(ctx)
	if gotSpaces != int64(expectedSpaces) {
		t.Errorf("spaces stat = %d, want %d", gotSpaces, expectedSpaces)
	}
	if gotUsers != int64(expectedUsers) {
		t.Errorf("verified_users stat = %d, want %d", gotUsers, expectedUsers)
	}
}

func TestStats_EnsureInitialized_NoOpWhenPresent(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	core.setStat(ctx, StatSpaces, 42)
	core.setStat(ctx, StatVerifiedUsers, 7)

	if err := core.EnsureStatsInitialized(ctx); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	// Existing values must NOT have been overwritten by recompute.
	got, _ := core.GetStat(ctx, StatSpaces)
	if got != 42 {
		t.Errorf("spaces stat overwritten: got %d, want 42 (preserved)", got)
	}
}

func TestStats_EnsureInitialized_SeedsWhenMissing(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Don't pre-set any stats. Create some real state so recompute has data.
	user, _ := core.CreateUser(ctx, "system", "seed-user", "Seed", "password123")
	core.AddVerifiedEmailDirect(ctx, user.Id, "seed@example.com")

	if err := core.EnsureStatsInitialized(ctx); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	got, _ := core.GetStat(ctx, StatVerifiedUsers)
	if got < 1 {
		t.Errorf("verified_users stat after seed = %d, want >= 1", got)
	}
}
