# @Mentions

## Overview

- Users can mention others with `@username` syntax in message bodies.
- Mentioned users receive a notification and a room-level mention indicator (more prominent than a regular unread marker).
- Only users who are members of the space can be mentioned. Invalid or non-member mentions are silently ignored.

## Autocomplete

- Typing `@` followed by at least one character triggers tab-completion in the composer.
- Matching is fuzzy against both login and display name, with prefix matches ranked higher.
- Pressing Tab completes the first match (appending a space). Pressing Tab again cycles through candidates.

## Rendering

- Valid mentions are highlighted in the message body. Self-mentions get additional styling.
- Mentions inside code blocks, pre-formatted text, and blockquotes are not styled.
- Invalid mentions (user not found or not a space member) are left as plain text.

## Notifications

- Mentioned users receive a persistent notification (bell icon) and a live mention event (room-level indicator).
- Self-mentions are skipped (you don't get notified for mentioning yourself).
- Mention notifications respect muting — if the user has muted the room, no notification is created.

## Interaction with Echo Events

- When a thread reply is echoed to the channel, the `mentionedUserIds` are copied to the echo event for independent notification delivery.
- Mentioned users are not notified twice (the original thread reply triggers the notification, not the echo).

## Permissions

- There are no separate mention permissions. Anyone who can post a message can mention any space member.
- No `@channel` or `@here` mentions exist — only individual user mentions.
