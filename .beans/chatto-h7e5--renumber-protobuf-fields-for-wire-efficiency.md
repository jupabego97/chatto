---
# chatto-h7e5
title: Renumber protobuf fields for wire efficiency
status: todo
type: task
priority: normal
created_at: 2026-02-16T13:50:27Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzk
parent: chatto-v29q
---

## Problem

Several high-frequency protobuf messages use field numbers above 15, which costs an extra byte per field per message on the wire. Since breaking changes are acceptable at this stage, we can renumber for efficiency.

### Fields to renumber

1. **`SpaceEvent.sequence_id = 9001`** → move to field 4-9 range
   - File: `proto/chatto/core/v1/space_event.proto:39`

2. **`MessagePostedEvent.sequence_id = 1000`** and **`event_id = 1001`** → move to 7-9 range
   - File: `proto/chatto/core/v1/space_event.proto:194-198`

3. **`MessageBody` fields**: `encrypted_body = 20`, `encryption_nonce = 21`, `attachments = 30`, `link_preview = 40` → renumber to 4-7
   - File: `proto/chatto/core/v1/models.proto:147-160`

### Also consider

4. **`Asset` oneof**: `NATSAsset nats = 2; S3Asset s3 = 1;` — swap so the more common NATS gets field 1
   - File: `proto/chatto/core/v1/models.proto:35-40`

5. **`Attachment.size`**: change from `int64` to `uint64` (file sizes are never negative)
   - File: `proto/chatto/core/v1/models.proto:129`

6. **`RoomMembership.space_id`**: remove redundant field (room IDs are globally unique)
   - File: `proto/chatto/core/v1/models.proto:56-60`

## Impact

Saves 2-3 bytes per message on the wire. Compounding effect at scale across all message events.

## Approach

This is a breaking change to the wire format. Since the project accepts breaking changes:
1. Renumber all fields
2. Run `mise codegen` to regenerate Go code
3. Run full test suite
4. Update ARCHITECTURE.md if it references specific field numbers
