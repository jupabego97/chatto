# ADR-031: Dual Asset Storage — NATS ObjectStore Default, S3 Optional

**Date:** 2026-03-01

> **Note:** Originally landed as ADR-021. Renumbered to ADR-031 on 2026-05-04 to resolve a numbering collision with [ADR-021 — Consolidate Instance and Space into a Single Server Concept](ADR-021-consolidate-instance-and-space-into-server.md). The decision content is unchanged.

## Context

Chatto stores binary assets: user avatars, space icons, and message attachments. The storage backend must work out of the box for small self-hosted instances but scale for larger deployments with terabytes of files.

The options are:

- **S3-only**: Simple, scalable, well-understood. But requires operators to provision an S3-compatible service, breaking the zero-dependency self-hosted goal.
- **NATS ObjectStore only**: Uses the existing NATS infrastructure, no external dependencies. But NATS isn't optimized for large blob storage and operators have no escape hatch.
- **Local filesystem**: Fast, simple. But not portable across processes in a clustered deployment.
- **Default to NATS, optional S3**: Zero dependencies for small instances, scalable storage for large ones.

## Decision

Support two storage backends with NATS ObjectStore as the default:

- **NATS ObjectStore** (default): Assets stored in JetStream-backed object store buckets. Works out of the box with zero configuration.
- **S3-compatible storage** (optional): When configured, new uploads go to S3. Existing NATS-stored assets continue to be served from NATS (dual-read, single-write-new).

Asset retrieval tries NATS first, then S3. This means switching to S3 doesn't require migrating existing assets — they remain accessible from NATS until the operator chooses to migrate them.

## Consequences

- **Zero-dependency default**: Small instances run entirely on embedded NATS. No S3 bucket to provision, no IAM credentials to configure.
- **Gradual migration path**: Operators can enable S3 at any time. New uploads go to S3; old assets remain in NATS. No downtime, no bulk migration required.
- **Dual-read overhead**: Asset retrieval may check two backends. In practice, after the transition period, most assets will be in S3 and the NATS lookup is a fast miss.
- **Deletion must handle both backends**: When deleting an asset, both NATS and S3 are checked independently. The code must tolerate "not found" from either backend.
- **NATS storage limits matter for NATS-only deployments**: Large instances using NATS-only storage will eventually hit practical limits (disk space, stream size). The S3 option is the escape hatch for this.
