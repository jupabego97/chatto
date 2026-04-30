# ADR-011: Instance Stats as a Counter Cache

**Date:** 2026-04-30

## Context

When we shipped instance-wide limits (`max_spaces`, `max_users`), enforcement was done with O(N) `ListKeysFiltered` scans on every gated mutation. Two problems with that:

1. **Inaccurate.** NATS KV `kv.Delete` writes a tombstone rather than removing the subject; with `History=1` the tombstone replaces the live value but the subject still has 1 message, so the scan needs to filter via the `KV-Operation: DEL` header. We do that correctly, but the scan still has to read every subject (live + tombstoned) to discriminate. Long-lived churned instances pay an ever-growing cost.
2. **Non-atomic.** "Read count, check, then write" is racy. Two concurrent signups at `count=N-1` can both pass the check and both succeed, briefly overshooting the limit. The original limits PR shipped with this caveat documented, and we removed a compensating check in `addVerifiedEmail` because it was a half-measure (still scan-based, also non-atomic).

We need O(1) reads, atomic enforcement, and a clean recovery story for when the cache drifts.

## Decision

Introduce **cached counters in the INSTANCE KV bucket**, treating them as a fast cache layer over authoritative KV state:

- One key per counter (`instance.stats.spaces`, `instance.stats.verified_users`). Per-counter keys avoid CAS contention between unrelated counters.
- **Value is just an ASCII decimal string in the KV entry.** `[]byte("42")`. No proto, no struct, no marshal/unmarshal — counters are just numbers and a wrapper type would be ceremony for nothing. Bonus: `nats kv get KV_INSTANCE instance.stats.spaces` returns a human-readable value.
- Atomic mutation via `IncrementStatIfBelow(name, max)`: a CAS loop that increments only if `count + 1 <= max`, returning `ErrLimitExceeded` otherwise. This is the gate.
- Drift recovery: **`RecomputeStats` runs on every startup.** Two `ListKeysFiltered` scans take milliseconds; doing it unconditionally means a fresh process always boots with truthful counters and there's no "is the counter present?" branching to maintain. If an operator suspects in-flight drift, the recovery procedure is to restart (optionally `nats kv del`-ing the misbehaving counter first for fine-grained control). No CLI command, no admin endpoint, no periodic background job — keep the surface small until something asks for more.

## Consequences

- **O(1) limit enforcement.** Every gated mutation does a single KV `Get` + CAS `Update`, not a stream scan.
- **Atomic limit gates.** `CreateSpace` and the verification-time `addVerifiedEmail` transition both use CAS-incrementing the counter against the limit. Two concurrent attempts at the boundary cannot both succeed.
- **Drift is possible but recoverable.** Counter mutations and the underlying KV writes are not transactionally linked. If the mutation fails after the increment, the helper rolls back; if the rollback itself fails, drift accumulates. Restart triggers a recompute. Drift is correctible with operator awareness, not invisible.
- **Two-gate user limit.** `CreateUser` reads `GetStat` (non-atomic, fast UX gate so users don't sign up only to be blocked at verification); `addVerifiedEmail` does the atomic CAS-increment on the unverified→verified transition. Subsequent emails on an already-verified user don't re-increment.
- **System spaces are excluded.** `CountSpaces` and the cached counter both ignore the DM system space, so user-facing limits reflect user-creatable spaces only.
- **Startup cost grows with instance size.** Two `ListKeysFiltered` scans on every boot. Negligible for typical deployments; if a Cloud tenant ever has 100k+ users, we revisit (e.g., move to a less frequent recompute or expose an explicit admin endpoint).

## Alternatives considered

- **Single bundled key with all counters.** Rejected: every counter increment fights for the same revision under CAS, causing false contention between unrelated mutations.
- **Separate KV bucket for stats.** Rejected: stats are persistent instance-level data with the same lifecycle as users/spaces. Reusing INSTANCE means no new backup-skip rules, no new replication concerns.
- **Server-side counts via `stream.Info(WithSubjectFilter)`.** Rejected: inflates by tombstones (KV `Delete` doesn't remove the subject). Documented in `CountSpaces` godoc.
- **Proto wrapper around the counter** with `count + updated_at + recomputed_at`. Rejected: forces codegen for what's literally one int. Restart time is the de facto "recomputed_at" — operators can read it from the server logs if they need it.
- **CLI command for recompute.** Rejected: with always-recompute-on-startup, the operator workflow is just "restart." If we ever need a no-restart path, we'll expose it as a GraphQL admin mutation.
- **Periodic background recompute.** Rejected: complexity without evidence of need.
