---
# chatto-wqpb
title: Graceful reconnection during rolling deployments
status: draft
type: feature
created_at: 2026-01-19T16:21:10Z
updated_at: 2026-01-19T16:21:10Z
---

Implement version-aware reconnection to ensure clients connect to new pods during K8s rolling deployments, avoiding repeated disconnections as old pods are cycled.

## Problem

During K8s rolling deployments:
1. Old pod terminates → client's WebSocket disconnects
2. Client reconnects → might hit another old pod still in the Service endpoints
3. That old pod terminates → client disconnects again
4. Cycle repeats until client is lucky enough to hit a new pod

This causes poor UX with multiple "Reconnecting..." banners in quick succession.

## Solution

Two-pronged approach:

### Part 1: Infrastructure - Proper Termination Sequencing

Add a `preStop` hook to the Kubernetes deployment that delays the SIGTERM, giving time for the pod to be removed from Service endpoints before connections are closed.

```yaml
spec:
  terminationGracePeriodSeconds: 60
  containers:
  - lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "sleep 10"]
```

This ensures reconnecting clients won't hit a pod that's actively terminating.

### Part 2: Application - Version-Aware Reconnection

Send the server's build version on WebSocket connection. Frontend tracks the highest version seen and, if it reconnects to an older version, disconnects and retries with exponential backoff.

#### Backend Changes

1. **Add version to GraphQL connection init response**
   - Include build version (git SHA or semver) in the connection acknowledgment
   - Could use `connectionParams` in the graphql-ws protocol or a custom field

2. **Add health/version endpoint** (optional, for debugging)
   - `GET /api/version` → `{ "version": "v0.0.30-abc123", "startedAt": "..." }`

#### Frontend Changes

1. **Track server version in connection state**
   - Store the version received on successful connection
   - Compare on reconnection

2. **Version mismatch handling**
   - If reconnected version < previously seen version:
     - Log warning
     - Disconnect immediately
     - Retry with backoff (1s, 2s, 4s... up to ~30s)
   - If reconnected version >= previously seen: normal operation

3. **UI feedback** (optional)
   - Could show "Updating..." instead of "Reconnecting..." when version mismatch detected
   - Gives users context that the reconnection is intentional

## Implementation Plan

- [ ] **Backend: Add version to WebSocket connection**
  - Determine where to inject version (connection_ack payload or custom message)
  - Add build version to CLI (embed via `-ldflags` at build time)
  - Include version in GraphQL WebSocket handshake

- [ ] **Frontend: Track and compare versions**
  - Update `graphqlClient.svelte.ts` to parse version from connection
  - Store highest-seen version in module state
  - Add version comparison logic on reconnection

- [ ] **Frontend: Implement version-aware reconnection**
  - Detect version mismatch
  - Force disconnect and retry with exponential backoff
  - Add logging for debugging

- [ ] **Documentation: Update deployment docs**
  - Add recommended `preStop` hook configuration
  - Document the version-aware reconnection behavior

## Open Questions

1. **Version format**: Should we use git SHA, semver, or both? Git SHA is unique per build; semver is more human-readable.

2. **Version source**: How do we get the "expected" version? Options:
   - Compare to previously-seen version (simpler, what we're proposing)
   - Fetch expected version from a separate endpoint (more complex, handles edge cases)

3. **Backoff strategy**: How aggressive should the retry backoff be? Need to balance quick reconnection with not hammering the service.

4. **graceful-ws library**: Does our current WebSocket setup (urql + graphql-ws) support custom reconnection logic, or do we need to hook in at a lower level?

## Related

- Kubernetes rolling update documentation
- graphql-ws connection protocol
- urql's WebSocket exchange reconnection behavior
