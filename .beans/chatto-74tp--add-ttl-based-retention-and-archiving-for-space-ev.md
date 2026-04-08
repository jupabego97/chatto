---
# chatto-74tp
title: Add TTL-based retention and archiving for space event streams
status: draft
type: feature
created_at: 2026-01-18T13:06:48Z
updated_at: 2026-01-18T13:06:48Z
---

Implement TTL-based retention on SPACE_{id}_EVENTS streams with archiving to long-term storage.

## Context

This is a companion to the event ID subject pattern change. With unique subjects per event, the subject index grows linearly. TTL-based retention bounds the index size to the retention period.

## Goals
- Bound memory usage by limiting hot data to N days (e.g., 30 days)
- Archive older messages to cold storage for compliance/history
- Keep the system responsive as message volume grows

## Tasks
- [ ] Research NATS stream MaxAge configuration
- [ ] Design archival format and storage (S3? local files?)
- [ ] Implement archive consumer that writes to cold storage
- [ ] Add configuration for retention period
- [ ] Test stream pruning behavior
- [ ] Document retention policy

## Open Questions
- What's the right default retention period? (7 days? 30 days?)
- Where should archives be stored? (Object store? Local disk?)
- Should archived messages be searchable?

## Related
- Depends on: chatto-mnqi (event ID subject patterns)