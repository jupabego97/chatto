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
- Value is a small protobuf (`InstanceStat`) holding `count`, `updated_at`, and `recomputed_at`. Metadata for free; future-proof if we ever want to store additional drift-debug info.
- Atomic mutation via `IncrementStatIfBelow(name, max)`: a CAS loop that increments only if `count + 1 <= max`, returning `ErrLimitExceeded` otherwise. This is the gate.
- Drift recovery is a first-class concern:
  - `RecomputeStats(ctx)` rebuilds counters from authoritative state via the existing `CountSpaces` / `CountVerifiedUsers` scans.
  - Runs automatically on startup if any well-known counter is missing (handles upgrades from instances that predate this system).
  - Exposed via `chatto stats recompute` for operator-initiated repair.

## Consequences

- **O(1) limit enforcement.** Every gated mutation does a single KV `Get` + CAS `Update`, not a stream scan.
- **Atomic limit gates.** `CreateSpace` and the verification-time `addVerifiedEmail` transition both use CAS-incrementing the counter against the limit. Two concurrent attempts at the boundary cannot both succeed.
- **Drift is possible.** Counter mutations and the underlying KV writes are not transactionally linked. If the mutation fails after the increment, the helper rolls back; if the rollback itself fails, drift accumulates. `RecomputeStats` is the safety valve. Drift is correctible but requires admin awareness — the `recomputed_at` field gives operators a signal.
- **Two-gate user limit.** `CreateUser` reads `GetStat` (non-atomic, fast UX gate so users don't sign up only to be blocked at verification); `addVerifiedEmail` does the atomic CAS-increment on the unverified→verified transition. Subsequent emails on an already-verified user don't re-increment.
- **System spaces are excluded.** `CountSpaces` and the cached counter both ignore the DM system space, so user-facing limits reflect user-creatable spaces only.
- **No periodic background recompute.** Manual + on-startup-if-missing is enough. Add periodic recompute only if drift becomes a measurable problem in practice.

## Alternatives considered

- **Single bundled key with all counters.** Rejected: every counter increment fights for the same revision under CAS, causing false contention between unrelated mutations.
- **Separate KV bucket for stats.** Rejected: stats are persistent instance-level data with the same lifecycle as users/spaces. Reusing INSTANCE means no new backup-skip rules, no new replication concerns.
- **Server-side counts via `stream.Info(WithSubjectFilter)`.** Rejected: inflates by tombstones (KV `Delete` doesn't remove the subject). Documented in `CountSpaces` godoc.
- **Periodic background recompute.** Rejected for now as complexity without evidence of need.
