# Message Editing & Deletion

## Editing

- Users can edit their own messages within a 3-hour window.
- Users with the `message.edit-any` permission can edit any message at any time.
- Only the message body text can be edited. Attachments can be removed separately but not edited.
- Edits use optimistic locking — if two edits race, one fails and the client must retry.

## Deletion

- Users can delete their own messages. Moderators with `message.delete-any` can delete any message.
- Deletion removes the message body and all attachments (GDPR-compliant). Edit and delete notifications are live-only events (not persisted to JetStream).
- Deleting an already-deleted message is a no-op (idempotent).
- Deleted messages display a "[Message deleted]" placeholder in the UI.

## Shared Message Body

- Thread reply echoes share the same message body as the original thread reply.
- Editing or deleting a message propagates to both the original and the echo automatically.
- Edit and delete live events carry a `messageEventId` (the event ID of the affected message); the frontend matches this against loaded events (by `e.id` or `echoOfEventId` for echoes) and refetches them.

## Related Operations

- Individual attachments and link previews can be removed from a message without deleting the whole message. These are author-only operations.

## Permissions

- `message.edit-own` — Edit one's own messages (within the 3-hour window)
- `message.edit-any` — Edit any user's messages (moderation, no time limit)
- `message.delete-own` — Delete one's own messages
- `message.delete-any` — Delete any user's messages (moderation)
