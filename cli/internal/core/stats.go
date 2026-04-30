package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
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

// GetStat returns the current value of a counter. Returns 0 if the counter
// hasn't been initialized yet (callers can treat that as the empty state).
func (c *ChattoCore) GetStat(ctx context.Context, name string) (int64, error) {
	stat, err := c.getStatProto(ctx, name)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return stat.Count, nil
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
			stat := &corev1.InstanceStat{Count: 1, UpdatedAt: timestamppb.Now()}
			data, err := proto.Marshal(stat)
			if err != nil {
				return 0, fmt.Errorf("marshal stat: %w", err)
			}
			if _, err := c.storage.instanceKV.Create(ctx, key, data); err != nil {
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

		var stat corev1.InstanceStat
		if err := proto.Unmarshal(entry.Value(), &stat); err != nil {
			return 0, fmt.Errorf("unmarshal stat: %w", err)
		}

		newCount := stat.Count + int64(delta)
		if newCount < 0 {
			newCount = 0 // floor decrements; never store negative
		}
		if delta > 0 && max >= 0 && newCount > max {
			return stat.Count, ErrLimitExceeded
		}

		stat.Count = newCount
		stat.UpdatedAt = timestamppb.Now()
		data, err := proto.Marshal(&stat)
		if err != nil {
			return 0, fmt.Errorf("marshal stat: %w", err)
		}
		if _, err := c.storage.instanceKV.Update(ctx, key, data, entry.Revision()); err != nil {
			if errors.Is(err, jetstream.ErrKeyExists) {
				continue // revision mismatch; another caller won the race
			}
			return 0, fmt.Errorf("update stat: %w", err)
		}
		return newCount, nil
	}
	return 0, fmt.Errorf("stat %q: exhausted %d CAS retries", name, statMaxRetries)
}

// getStatProto fetches the raw InstanceStat. Internal helper.
func (c *ChattoCore) getStatProto(ctx context.Context, name string) (*corev1.InstanceStat, error) {
	entry, err := c.storage.instanceKV.Get(ctx, statKey(name))
	if err != nil {
		return nil, err
	}
	var stat corev1.InstanceStat
	if err := proto.Unmarshal(entry.Value(), &stat); err != nil {
		return nil, fmt.Errorf("unmarshal stat: %w", err)
	}
	return &stat, nil
}

// setStat overwrites a counter to the given value. Used by RecomputeStats to
// re-establish truth from authoritative state. Updates RecomputedAt so admins
// can see when the counter was last reconciled.
func (c *ChattoCore) setStat(ctx context.Context, name string, count int64) error {
	now := timestamppb.Now()
	stat := &corev1.InstanceStat{
		Count:        count,
		UpdatedAt:    now,
		RecomputedAt: now,
	}
	data, err := proto.Marshal(stat)
	if err != nil {
		return fmt.Errorf("marshal stat: %w", err)
	}
	if _, err := c.storage.instanceKV.Put(ctx, statKey(name), data); err != nil {
		return fmt.Errorf("put stat: %w", err)
	}
	return nil
}

// RecomputeStats scans authoritative state and overwrites the well-known
// counters with truth. Run on startup to seed counters on instances upgraded
// from versions that predate this system, and exposed via `chatto stats
// recompute` for manual repair if drift is suspected.
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

// EnsureStatsInitialized seeds counters on first run by recomputing if any of
// the well-known stat keys are missing. Idempotent and cheap when stats already
// exist.
func (c *ChattoCore) EnsureStatsInitialized(ctx context.Context) error {
	for _, name := range []string{StatSpaces, StatVerifiedUsers} {
		_, err := c.storage.instanceKV.Get(ctx, statKey(name))
		if err == nil {
			continue
		}
		if !errors.Is(err, jetstream.ErrKeyNotFound) {
			return fmt.Errorf("check stat %q: %w", name, err)
		}
		// At least one is missing — recompute the lot.
		c.logger.Info("instance stats missing; recomputing from authoritative state", "missing", name)
		return c.RecomputeStats(ctx)
	}
	return nil
}
