package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/nats-io/nats.go/jetstream"
)

// Names of the well-known instance counters. Add new ones here so they're
// declared centrally and picked up by RecomputeStats.
const (
	StatSpaces        = "spaces"
	StatVerifiedUsers = "verified_users"
)

// statKey returns the KV key for a counter. Stored in the INSTANCE bucket.
// Format: instance.stats.{name}.
func statKey(name string) string {
	return "instance.stats." + name
}

// statMaxRetries caps the CAS retry budget for stat updates. Same reasoning as
// the rate limiter: under contention, the slowest caller can need ~N-1 retries.
// 32 covers realistic contention against a single counter without busy-looping.
const statMaxRetries = 32

// Counter values are stored as ASCII decimal bytes in the KV entry — `nats kv
// get KV_INSTANCE instance.stats.spaces` returns "42". Plain bytes were chosen
// over a proto wrapper because the value is just a number; no need for a
// marshal/unmarshal step or codegen to add a new counter.

// GetStat returns the current value of a counter. Returns 0 if the counter
// hasn't been initialized yet (callers can treat that as the empty state).
func (c *ChattoCore) GetStat(ctx context.Context, name string) (int64, error) {
	entry, err := c.storage.instanceKV.Get(ctx, statKey(name))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("get stat: %w", err)
	}
	count, err := strconv.ParseInt(string(entry.Value()), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse stat %q: %w", name, err)
	}
	return count, nil
}

// IncrementStat increments a counter by 1 and returns the new count. CAS-loops
// on revision mismatch under contention.
func (c *ChattoCore) IncrementStat(ctx context.Context, name string) (int64, error) {
	return c.adjustStat(ctx, name, +1, -1 /* no limit */)
}

// DecrementStat decrements a counter by 1, floored at 0 (never goes negative —
// drift would otherwise propagate forever as a negative value). Returns the
// new count.
func (c *ChattoCore) DecrementStat(ctx context.Context, name string) (int64, error) {
	return c.adjustStat(ctx, name, -1, -1 /* no limit */)
}

// IncrementStatIfBelow atomically increments iff (current + 1 <= max). Returns
// ErrLimitExceeded if at the limit. A negative max disables the check (useful
// for callers that conditionally enforce a limit). Pass max == 0 to lock.
//
// This is the atomic limit gate — preferred over GetStat-then-Increment because
// it closes the race window between check and increment.
func (c *ChattoCore) IncrementStatIfBelow(ctx context.Context, name string, max int64) (int64, error) {
	return c.adjustStat(ctx, name, +1, max)
}

// adjustStat is the shared CAS loop for Increment/Decrement/IncrementStatIfBelow.
// delta is +1 or -1; max < 0 disables the limit check.
func (c *ChattoCore) adjustStat(ctx context.Context, name string, delta int, max int64) (int64, error) {
	key := statKey(name)
	for attempt := 0; attempt < statMaxRetries; attempt++ {
		entry, err := c.storage.instanceKV.Get(ctx, key)

		if errors.Is(err, jetstream.ErrKeyNotFound) {
			// Counter doesn't exist yet. For increments we Create at 1; for
			// decrements we leave it absent (a missing counter is "0"; -1 isn't
			// meaningful here). For IncrementStatIfBelow, refuse if max == 0.
			if delta < 0 {
				return 0, nil
			}
			if max == 0 {
				return 0, ErrLimitExceeded
			}
			if _, err := c.storage.instanceKV.Create(ctx, key, []byte("1")); err != nil {
				if errors.Is(err, jetstream.ErrKeyExists) {
					continue // raced with another caller; retry through the Update path
				}
				return 0, fmt.Errorf("create stat: %w", err)
			}
			return 1, nil
		}
		if err != nil {
			return 0, fmt.Errorf("get stat: %w", err)
		}

		current, err := strconv.ParseInt(string(entry.Value()), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse stat %q: %w", name, err)
		}

		newCount := current + int64(delta)
		if newCount < 0 {
			newCount = 0 // floor decrements; never store negative
		}
		if delta > 0 && max >= 0 && newCount > max {
			return current, ErrLimitExceeded
		}

		if _, err := c.storage.instanceKV.Update(ctx, key, []byte(strconv.FormatInt(newCount, 10)), entry.Revision()); err != nil {
			if errors.Is(err, jetstream.ErrKeyExists) {
				continue // revision mismatch; another caller won the race
			}
			return 0, fmt.Errorf("update stat: %w", err)
		}
		return newCount, nil
	}
	return 0, fmt.Errorf("stat %q: exhausted %d CAS retries", name, statMaxRetries)
}

// setStat overwrites a counter to the given value. Used by RecomputeStats to
// re-establish truth from authoritative state.
func (c *ChattoCore) setStat(ctx context.Context, name string, count int64) error {
	if _, err := c.storage.instanceKV.Put(ctx, statKey(name), []byte(strconv.FormatInt(count, 10))); err != nil {
		return fmt.Errorf("put stat: %w", err)
	}
	return nil
}

// RecomputeStats scans authoritative state and overwrites the well-known
// counters with truth. Called from NewChattoCore on every startup so a fresh
// process always boots with counters that match reality — drift recovery is
// "restart the server" (or `nats kv del` for the misbehaving counter and
// restart for finer control). Two ListKeysFiltered scans; cheap.
func (c *ChattoCore) RecomputeStats(ctx context.Context) error {
	spaces, err := c.CountSpaces(ctx)
	if err != nil {
		return fmt.Errorf("count spaces: %w", err)
	}
	if err := c.setStat(ctx, StatSpaces, int64(spaces)); err != nil {
		return err
	}

	users, err := c.CountVerifiedUsers(ctx)
	if err != nil {
		return fmt.Errorf("count verified users: %w", err)
	}
	if err := c.setStat(ctx, StatVerifiedUsers, int64(users)); err != nil {
		return err
	}

	c.logger.Info("recomputed instance stats", "spaces", spaces, "verified_users", users)
	return nil
}
