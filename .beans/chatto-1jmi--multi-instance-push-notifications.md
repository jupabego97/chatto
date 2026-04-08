---
# chatto-1jmi
title: Multi-instance push notifications
status: todo
type: task
created_at: 2026-03-26T15:14:12Z
updated_at: 2026-03-26T15:14:12Z
parent: chatto-e88o
---

Push notification subscribe/unsubscribe is hardcoded to originClient in pushNotifications.ts. Need to support push subscriptions per instance so remote instance notifications work.

## Current State

- `subscribe()` (line ~137) and `unsubscribe()` (line ~179) in `pushNotifications.ts` use `graphqlClientManager.originClient.client` hardcoded
- `PushNotificationSetup.svelte` only checks `originInstanceState?.pushNotificationsEnabled`
- Only the origin instance can manage push subscriptions; remote instance notifications don't trigger push

## Needs Decision

- One push subscription per instance, or aggregate through origin?
- Service worker is tied to origin — how should it route remote instance notifications?

## Files to Modify

- `frontend/src/lib/notifications/pushNotifications.ts` — accept instanceId, use correct client
- `frontend/src/lib/components/PushNotificationSetup.svelte` — set up push for all instances
