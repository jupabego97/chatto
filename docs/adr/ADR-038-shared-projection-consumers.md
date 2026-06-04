# ADR-038: Shared Projection Consumers with Projection-Scoped Failure Isolation

**Date:** 2026-06-04

## Context

ADR-033 introduced event-sourced state with in-memory projections, and ADR-034 put all durable domain facts into one `EVT` stream. The initial projection runtime kept the mechanics intentionally simple: each `Projector` owned one NATS `OrderedConsumer` over its projection's declared `Subjects()`.

That shape was correct when only a small number of projections existed. It is becoming wasteful as more read models migrate to EVT: several projections subscribe to identical or overlapping subject filters, so one process may replay the same slice of the stream many times during boot. The code still needs the useful parts of the current model:

- `WaitForSeq` and projection readiness are per read model.
- A broken projection must not silently corrupt its state.
- Future snapshot/rebuild work should be able to restart one projection without redesigning all event consumers.

The next runtime shape needs to reduce consumer fanout while keeping projection lifecycle isolated.

## Decision

Refactor the projection runtime so NATS consumers are owned by a shared projection manager, not by each individual projection handle.

- **Group by canonical subject filter set.** Normalize each projection's `Subjects()` by sorting and deduplicating the filters. Projections with the same canonical filter set share one NATS `OrderedConsumer`. Projections with different filter sets keep separate consumers.
- **Keep projection handles per projection.** The existing `Projector` remains the handle used by `ChattoCore`, admin diagnostics, and write paths. It owns per-projection state such as `LastSeq`, `WaitForSeq`, started/current readiness, declared subjects, and failure state, but it no longer needs to own a JetStream consumer directly.
- **Dispatch in stream order within each group.** The manager consumes one message at a time per filter group, unmarshals the durable `corev1.Event` once, and applies it to every healthy projection in that group. The `Projection.Apply(event, seq)` contract stays unchanged.
- **Isolate projection failures.** If one projection's `Apply` returns an error, only that projection is marked failed at that stream sequence. Its waiters receive `ErrProjectionFailed` for that sequence or later. Other projections in the same consumer group continue applying subsequent events and advancing their own sequence.
- **Preserve per-projection catch-up semantics.** `CurrentTargetSeq` continues to compute the highest stream sequence matching the individual projection's subjects, not the broader group as a policy shortcut. `WaitForSeq` remains per projection.
- **Leave restart/rebuild policy explicit.** V1 does not automatically replay a failed projection. A failed projection remains unhealthy until process restart or a future operator/API-driven restart. The new boundary is intentionally shaped so a later `Restart` or `Rebuild` can replay one projection from snapshot or stream history without changing projection implementations.

This ADR updates the projection-runtime consequences of ADR-033 and ADR-034. It does not change the `EVT` subject layout, the GraphQL schema, event payloads, or the `Projection` interface.

## Consequences

- **Startup replay cost scales with filter groups, not projection count.** Multiple projections that need the same slice of `EVT` share the NATS consumer and decode path.
- **Projection state remains independently observable.** Admin views can still report per-projection subjects, lag, entry counts, estimated bytes, started/current state, and failure status.
- **Failure behavior stays conservative.** A projection crash is visible and contained. The system avoids letting a failed projection advance as if it had applied an event, while unrelated projections keep serving current reads.
- **The manager becomes a small internal scheduler.** It must own grouping, message dispatch, per-group consumer lifecycle, and error logging. That is more runtime structure than the original one-projector-one-consumer loop, but it is still local to the internal events package.
- **Broader multiplexing remains deferred.** Grouping by exact canonical subject set is the v1 rule. A single broad `evt.>` consumer with per-projection subject matching may become attractive later, but it has different backpressure and over-delivery tradeoffs and is not adopted here.
- **Snapshot orchestration is still separate.** The projection interface already accommodates snapshots; this ADR only preserves the lifecycle boundary needed for future restart/rebuild work.
