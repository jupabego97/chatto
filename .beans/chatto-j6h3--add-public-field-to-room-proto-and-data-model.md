---
# chatto-j6h3
title: Add public field to Room proto and data model
status: todo
type: task
priority: normal
created_at: 2026-02-10T08:05:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzk
parent: chatto-2quz
---

Add a \`public\` boolean field to the Room protobuf message and propagate through codegen.

## Implementation

### Proto change
- Add \`bool public = 7;\` to \`Room\` message in \`proto/chatto/core/v1/models.proto\`
- Run \`mise codegen\` to regenerate Go and frontend types

### Core changes
- Update \`CreateRoom\` in \`cli/internal/core/rooms.go\` to accept an optional \`public\` parameter
- The field is immutable after creation: \`UpdateRoom\` must NOT accept or modify the \`public\` field
- Add validation: if \`public\` is set, ensure it's respected during room creation but rejected on update

### GraphQL schema
- Add \`public: Boolean!\` field to the \`Room\` type
- Add \`public: Boolean\` (optional, defaults to false) to \`CreateRoomInput\`
- Do NOT add it to \`UpdateRoomInput\` (immutable after creation)
- Add \`viewerCanCreatePublicRoom\` field or reuse \`rooms.manage\` permission for this

### Permission check
- Only users with \`rooms.manage\` (or a new dedicated permission) should be able to create public rooms
- Regular users with \`rooms.create\` can only create private rooms

## Key files
- \`proto/chatto/core/v1/models.proto\`
- \`cli/internal/core/rooms.go\` (CreateRoom, UpdateRoom)
- \`cli/internal/graph/*.graphqls\` (schema)
- \`cli/internal/graph/room.resolvers.go\`

## Tests
- Unit test: CreateRoom with public=true stores the flag
- Unit test: UpdateRoom rejects attempts to change public flag
- Unit test: non-admin cannot create public rooms
