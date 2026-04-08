---
# chatto-bwc7
title: Add conditional auth bypass for public room queries
status: todo
type: task
priority: normal
created_at: 2026-02-10T08:05:29Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzs
parent: chatto-2quz
blocked_by:
    - chatto-j6h3
---

Modify GraphQL resolvers to allow unauthenticated access to public rooms. This is the core backend change that enables the feature.

## Implementation

### New auth helper
Add a helper in \`cli/internal/graph/authz.go\` that checks public room access:

\`\`\`go
// requireRoomAccess returns the authenticated user (may be nil for public rooms).
// For private rooms, requires authentication + room membership.
// For public rooms, allows unauthenticated access.
func requireRoomAccess(ctx context.Context, core *core.ChattoCore, spaceID, roomID string) (*corev1.User, error) {
    room, err := core.GetRoom(ctx, spaceID, roomID)
    if err != nil { return nil, err }
    
    if room.Public {
        // Public room: allow access, return user if authenticated (nil if not)
        user := auth.ForContext(ctx)
        return user, nil
    }
    
    // Private room: require auth + membership (existing behavior)
    return requireRoomMember(ctx, core, spaceID, roomID)
}
\`\`\`

### Resolvers to modify
These resolvers currently call \`requireRoomMember\` and need the conditional path:

1. **\`Room\` query** (\`query.resolvers.go\`) — fetch room details
2. **\`RoomEvents\` query** (\`query.resolvers.go\`) — fetch message history
3. **\`RoomEvent\` query** (\`query.resolvers.go\`) — fetch single event
4. **\`ThreadEvents\` query** — fetch thread replies

### Field-level filtering for public access
When \`user == nil\` (unauthenticated public access), restrict field resolvers:

- \`Room.members\`: return empty list or error (don't expose member list publicly)
- \`Room.hasUnread\`: return false (no user to track)
- \`Message.reactions\`: show counts only, not who reacted (privacy)
- Author info: show display name and avatar only (no email)

### What stays auth-required
All mutations remain auth-required, no changes needed:
- \`postMessage\`, \`addReaction\`, \`deleteMessage\`, etc.
- All subscriptions remain auth-required

## Key files
- \`cli/internal/graph/authz.go\` (new helper)
- \`cli/internal/graph/query.resolvers.go\` (Room, RoomEvents, RoomEvent)
- \`cli/internal/graph/room.resolvers.go\` (field resolvers)
- \`cli/internal/graph/message.resolvers.go\` (field filtering)

## Tests
- Query public room without auth: succeeds, returns room + events
- Query public room with auth: succeeds (authenticated users can also view)
- Query private room without auth: returns error
- Query private room without membership: returns error (unchanged)
- Public room member list: not exposed to unauthenticated users
- Public room message body decryption works for unauthenticated readers
