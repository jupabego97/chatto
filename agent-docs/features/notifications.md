# Notifications

## Overview

- Chatto has a persistent notification system with a bell icon and notification center.
- Notifications are created for: DM messages, @mentions, replies to your messages, thread replies (for followers), and all messages in rooms where the user has set ALL_MESSAGES notification level.
- Notifications have a 90-day TTL and auto-expire.
- Dismissing a notification removes it from all devices (cross-device sync via live events).

## Notification Levels

Users can set a notification level per space and per room:

- **DEFAULT** — Inherit from parent (room inherits space, space inherits system default of NORMAL)
- **MUTED** — Suppress all notifications AND unread markers. The room doesn't appear unread in the sidebar.
- **NORMAL** — Standard behavior. Notifications for mentions, DMs, and thread replies only.
- **ALL_MESSAGES** — Like NORMAL, plus a notification for every root message in the room.

Muting a room suppresses all notification types, including @mentions.

## Thread Follow Notifications

- Users are automatically followed to a thread when they post in it.
- Thread followers receive notifications for new replies (except their own).
- Users can manually follow/unfollow threads.
- Thread follow notifications respect muting — if the room is muted, no thread notifications are created.

## Real-Time Sync

- When a notification is created or dismissed, a live event is published for cross-tab/cross-device sync.
- The frontend plays a notification sound and updates the badge count in real time.
- Separate live events exist for room-level mention indicators (more prominent than general unread).

## Interaction with Push Notifications

- Push notifications are sent via a callback when a persistent notification is created.
- If no persistent notification is created (e.g., room is muted), no push is sent either.
- Dismissing a notification sends a "dismiss" push to other devices to close the system notification.
