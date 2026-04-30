package core

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimitCheck_AllowsUnderLimit(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := 0; i < 3; i++ {
		allowed, _, err := core.RateLimitCheck(ctx, "test.allow", "key-a", 3, time.Minute)
		if err != nil {
			t.Fatalf("call %d unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("call %d should be allowed under limit", i)
		}
	}
}

func TestRateLimitCheck_DeniesAtLimit(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := 0; i < 3; i++ {
		if allowed, _, _ := core.RateLimitCheck(ctx, "test.deny", "key-b", 3, time.Minute); !allowed {
			t.Fatalf("setup call %d should be allowed", i)
		}
	}

	allowed, retryAfter, err := core.RateLimitCheck(ctx, "test.deny", "key-b", 3, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Errorf("4th call should be denied")
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Errorf("retryAfter should be 0..1m, got %v", retryAfter)
	}
}

func TestRateLimitCheck_ScopesAreIndependent(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := 0; i < 2; i++ {
		core.RateLimitCheck(ctx, "test.scope-a", "shared-key", 2, time.Minute)
	}
	// scope-a is now exhausted, but scope-b should still allow this key.
	allowed, _, _ := core.RateLimitCheck(ctx, "test.scope-b", "shared-key", 2, time.Minute)
	if !allowed {
		t.Errorf("different scope should not share counters")
	}
}

func TestRateLimitCheck_KeysAreIndependent(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := 0; i < 2; i++ {
		core.RateLimitCheck(ctx, "test.keys", "ip-1", 2, time.Minute)
	}
	allowed, _, _ := core.RateLimitCheck(ctx, "test.keys", "ip-2", 2, time.Minute)
	if !allowed {
		t.Errorf("different key in same scope should not share counter")
	}
}

func TestRateLimitCheck_WindowExpiry(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Use a tiny window so we can wait it out without slowing the test.
	window := 200 * time.Millisecond

	for i := 0; i < 2; i++ {
		core.RateLimitCheck(ctx, "test.win", "key-c", 2, window)
	}
	if allowed, _, _ := core.RateLimitCheck(ctx, "test.win", "key-c", 2, window); allowed {
		t.Fatalf("3rd call within window should be denied")
	}

	time.Sleep(window + 50*time.Millisecond)

	if allowed, _, _ := core.RateLimitCheck(ctx, "test.win", "key-c", 2, window); !allowed {
		t.Errorf("call after window expiry should be allowed")
	}
}

func TestRateLimitCheck_ConcurrentCASIsCorrect(t *testing.T) {
	// Hammer the same (scope,key) from many goroutines and verify the total number
	// of "allowed" responses equals the configured max — i.e. CAS retries don't
	// allow over-shoot.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	const max = 10
	const goroutines = 50

	var wg sync.WaitGroup
	allowedCount := 0
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _, _ := core.RateLimitCheck(ctx, "test.concurrent", "race-key", max, time.Minute)
			if allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowedCount != max {
		t.Errorf("expected exactly %d calls allowed under contention, got %d", max, allowedCount)
	}
}

func TestRateLimitCheck_ZeroMaxAllowsAll(t *testing.T) {
	// Sentinel: max <= 0 disables the limit so callers can pass 0 to opt out.
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	for i := 0; i < 100; i++ {
		if allowed, _, _ := core.RateLimitCheck(ctx, "test.disabled", "k", 0, time.Minute); !allowed {
			t.Fatalf("max=0 should allow unconditionally, denied on call %d", i)
		}
	}
}
