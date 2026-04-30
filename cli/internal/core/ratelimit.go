package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// rateLimitState is what we store in the RATELIMITS bucket. Compact field names
// keep the JSON small since this is a hot path with many short-lived entries.
type rateLimitState struct {
	Count       int   `json:"c"`
	WindowStart int64 `json:"w"` // unix milliseconds (sub-second precision matters for short windows in tests)
}

// rateLimitMaxRetries caps the CAS retry loop. With N concurrent callers all racing
// to increment the same counter, the slowest caller can need up to N-1 retries (each
// round, only one Update wins; losers re-Get and try again). 32 covers realistic
// burst sizes against a single rate-limit key without busy-looping forever; in
// production, typical contention is far below this since the limit itself caps the
// per-key burst rate.
const rateLimitMaxRetries = 32

// rateLimitKey builds the KV key for a (scope, key) pair. The key portion is hashed
// so callers can pass arbitrary input (IPs, emails) without worrying about NATS
// subject character restrictions or PII leakage in stream subjects.
func rateLimitKey(scope, key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%s.%s", scope, hex.EncodeToString(hash[:]))
}

// RateLimitCheck enforces a fixed-window rate limit and returns whether the call
// is allowed. On allow it atomically increments the per-(scope,key) counter; on
// deny it returns the time until the current window closes so the caller can set
// a Retry-After header.
//
// scope namespaces buckets (e.g. "register.ip", "register.email"); key is the
// per-caller identifier (the IP or email). max is the maximum number of allowed
// calls per window.
//
// Implementation: stateful counter with optimistic CAS. Each call reads the
// current state, resets if the window has expired, checks the count, and either
// updates (existing entry) or creates (fresh window) with KeyTTL=2*window so old
// state ages out automatically. Retries up to rateLimitMaxRetries times on
// revision mismatch.
//
// Returns allowed=false with a non-nil err only when the underlying KV operation
// fails. For consistency under contention the caller should treat err != nil
// as "fail open" or "fail closed" depending on how strict the gate needs to be.
func (c *ChattoCore) RateLimitCheck(ctx context.Context, scope, key string, max int, window time.Duration) (allowed bool, retryAfter time.Duration, err error) {
	if max <= 0 || window <= 0 {
		return true, 0, nil
	}

	kvKey := rateLimitKey(scope, key)
	now := time.Now()
	nowMs := now.UnixMilli()
	// NATS rejects per-message TTL below 1 second, so clamp the floor for safety.
	// The window logic itself is window-precision; this only governs how soon stale
	// entries are GC'd from KV.
	windowEntryTTL := 2 * window
	if windowEntryTTL < time.Second {
		windowEntryTTL = time.Second
	}

	for attempt := 0; attempt < rateLimitMaxRetries; attempt++ {
		entry, getErr := c.storage.ratelimitKV.Get(ctx, kvKey)

		// Fresh window — either the key doesn't exist or the window has rolled over.
		// In both cases we Create a new entry. If another caller wins the race,
		// we retry through the Update path.
		if errors.Is(getErr, jetstream.ErrKeyNotFound) {
			state := rateLimitState{Count: 1, WindowStart: nowMs}
			data, _ := json.Marshal(state)
			if _, createErr := c.storage.ratelimitKV.Create(ctx, kvKey, data, jetstream.KeyTTL(windowEntryTTL)); createErr != nil {
				if errors.Is(createErr, jetstream.ErrKeyExists) {
					continue
				}
				return false, 0, fmt.Errorf("rate limit create failed: %w", createErr)
			}
			return true, 0, nil
		}
		if getErr != nil {
			return false, 0, fmt.Errorf("rate limit get failed: %w", getErr)
		}

		var state rateLimitState
		if unmarshalErr := json.Unmarshal(entry.Value(), &state); unmarshalErr != nil {
			// Corrupted entry — treat as fresh window and overwrite.
			state = rateLimitState{}
		}

		windowEnd := time.UnixMilli(state.WindowStart).Add(window)
		if !now.Before(windowEnd) {
			// Window expired; reset.
			state = rateLimitState{Count: 1, WindowStart: nowMs}
		} else {
			if state.Count >= max {
				return false, time.Until(windowEnd), nil
			}
			state.Count++
		}

		data, _ := json.Marshal(state)
		if _, updateErr := c.storage.ratelimitKV.Update(ctx, kvKey, data, entry.Revision()); updateErr != nil {
			if errors.Is(updateErr, jetstream.ErrKeyExists) {
				// Revision mismatch — another caller incremented first; retry.
				continue
			}
			return false, 0, fmt.Errorf("rate limit update failed: %w", updateErr)
		}
		return true, 0, nil
	}

	return false, 0, fmt.Errorf("rate limit check exhausted retries for %s/%s", scope, key)
}
