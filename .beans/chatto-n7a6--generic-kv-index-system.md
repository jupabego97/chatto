---
# chatto-n7a6
title: Generic KV Index System
status: draft
type: feature
created_at: 2026-01-23T09:17:05Z
updated_at: 2026-01-23T09:17:05Z
---

## Overview

Build a generic, rebuildable index system for NATS KV that supports:
- Declarative index definitions
- On-demand rebuilds (admin-triggered)
- Leader election for clustered deployments
- Admin UI for monitoring and triggering rebuilds

## Current Indexes

| Index | Source | Key Function | Notes |
|-------|--------|--------------|-------|
| `user_by_login.{login}` | `user.*` | `strings.ToLower(login)` | Case-insensitive |
| `user_by_email.{hash}` | verified emails | `sha256(strings.ToLower(email))` | Hash for valid NATS chars |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     IndexRegistry                        │
│  - RegisterIndex(def IndexDefinition)                   │
│  - GetIndex(name) IndexDefinition                       │
│  - ListIndexes() []IndexDefinition                      │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│                    IndexRebuilder                        │
│  - RebuildIndex(ctx, name) error                        │
│  - RebuildAll(ctx) error                                │
│  - GetStatus(name) IndexStatus                          │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   LeaderElection                         │
│  - AcquireLease(ctx) (Lease, error)                     │
│  - IsLeader() bool                                      │
│  - Release()                                            │
└─────────────────────────────────────────────────────────┘
```

## Package Structure

```
cli/internal/core/indexing/
├── index.go       # IndexDefinition type
├── registry.go    # Index registry
├── rebuilder.go   # Rebuild logic
├── leader.go      # Leader election
└── *_test.go      # Tests for each
```

## Key Types

### Index Definition

```go
type IndexDefinition struct {
    Name           string
    Description    string
    SourcePattern  string                                          // KV key pattern, e.g. "user.*"
    SourceFilter   func(key string) bool                          // Filter to relevant keys
    KeyExtractor   func(key string, value []byte) (string, error) // Extract index key
    ValueExtractor func(key string, value []byte) ([]byte, error) // Extract index value
}
```

### Rebuild Result

```go
type RebuildResult struct {
    IndexName       string
    Created         int
    Skipped         int  // Already existed with correct value
    Updated         int  // Existed with wrong value
    Conflicts       int  // Duplicate keys from different sources
    Errors          int
    Duration        time.Duration
    ConflictDetails []ConflictDetail
}
```

## Leader Election

KV-based lease using JetStream:
- Lease key: `_system.leader.index_rebuild`
- Lease TTL: 30 seconds
- Refresh interval: 10 seconds
- If lease expires (node crash), other nodes can acquire after TTL

## Admin API

```graphql
extend type AdminQueries {
    indexes: [IndexInfo!]!
}

extend type AdminMutations {
    rebuildIndex(name: String!): RebuildResult!
    rebuildAllIndexes: [RebuildResult!]!
}
```

## Conflict Handling

When rebuilding, if two source records map to the same index key:
1. **First-wins**: First record encountered gets the index entry
2. **Log conflict**: Record both source keys and the conflicting index key
3. **Include in result**: `RebuildResult.Conflicts` count + details
4. **Don't fail**: Continue processing other records

Admin can then manually resolve (rename login, delete duplicate account, etc.)

## Work Breakdown

- [ ] Index types (`indexing/index.go`) - Define `IndexDefinition`, `ErrSkipRecord`, support for multi-key extraction
- [ ] Registry (`indexing/registry.go`) - Thread-safe registry with Register/Get/List
- [ ] Leader election (`indexing/leader.go`) - KV-based lease with TTL and refresh
- [ ] Rebuilder (`indexing/rebuilder.go`) - Core rebuild logic with conflict handling
- [ ] Tests (`indexing/*_test.go`) - Unit + integration tests for all components
- [ ] Core integration (`core/core.go`, `core/indexes.go`) - Wire up indexing system, define concrete indexes
- [ ] GraphQL API (`graph/admin.graphqls`, `graph/admin.resolvers.go`) - Admin queries and mutations
- [ ] Frontend UI (`frontend/.../admin/indexes/`) - Admin page with rebuild buttons

## Verification

1. `mise test-cli` - All tests pass
2. Start server, navigate to `/chat/admin/indexes`
3. Click "Rebuild All" - see results
4. Verify indexes are populated correctly
5. Test with multiple nodes (if possible) to verify leader election
