---
# chatto-623c
title: 'PWA: Add app shortcuts to manifest'
status: draft
type: feature
created_at: 2026-02-07T19:41:44Z
updated_at: 2026-02-07T19:41:44Z
---

## Summary

Add `shortcuts` to Chatto's PWA manifest so users can long-press the app icon (mobile) or right-click it (desktop taskbar/dock) to get quick-access actions like jumping to DMs, browsing spaces, or opening notifications.

This is a low-effort, high-polish improvement — it's purely a manifest change plus ensuring the target routes work when opened directly.

## Background

The [Shortcuts API](https://developer.chrome.com/docs/capabilities/shortcuts) allows PWAs to expose up to 4 app shortcuts (platform-dependent, 4 is the common max). Each shortcut has a name, URL, and optional icon. The OS displays these in the app icon context menu.

**Browser support:** Chromium-based browsers (Chrome 96+, Edge) on Android, Windows, macOS, Linux. Safari does not support it yet.

## Requirements

### Manifest Changes

Add `shortcuts` array to `manifest.webmanifest`:

```json
{
  "shortcuts": [
    {
      "name": "Direct Messages",
      "short_name": "DMs",
      "url": "/chat/dm",
      "description": "Open your direct messages",
      "icons": [{ "src": "/icons/shortcut-dm.png", "sizes": "96x96" }]
    },
    {
      "name": "Notifications",
      "short_name": "Notifications",
      "url": "/chat/notifications",
      "description": "View your notifications",
      "icons": [{ "src": "/icons/shortcut-notifications.png", "sizes": "96x96" }]
    },
    {
      "name": "Browse Spaces",
      "short_name": "Spaces",
      "url": "/chat/spaces",
      "description": "Discover and join spaces",
      "icons": [{ "src": "/icons/shortcut-spaces.png", "sizes": "96x96" }]
    }
  ]
}
```

### Tasks

- [ ] Decide on the shortcut entries (3-4 max) — DMs, Notifications, Browse Spaces are strong candidates
- [ ] Add `shortcuts` array to `manifest.webmanifest`
- [ ] Generate shortcut icons (96x96px, monochrome or simple colored icons)
  - Option A: Generate from a set of SVG source icons using the existing `sharp` pipeline
  - Option B: Use the same app icon for all shortcuts (simpler, still works)
- [ ] Verify that each target URL works correctly when opened directly (not just via in-app navigation)
  - `/chat/dm` — should load DM list (may require auth redirect)
  - `/chat/notifications` — should load notifications page
  - `/chat/spaces` — should load space browser
- [ ] Test on Android (Chrome) and desktop (Chrome/Edge) to confirm shortcuts appear

### Icon Considerations

Shortcut icons are optional but recommended. They should be:
- 96x96px minimum
- Visually distinct from each other
- Simple enough to be legible at small sizes
- Can be monochrome (the OS may apply its own masking/tinting on some platforms)

If generating custom icons feels like too much scope, using the app icon for all shortcuts is a fine starting point — the text labels differentiate them.

## Non-Goals

- Dynamic shortcuts (changing based on user's pinned spaces, recent rooms) — not possible with static manifest
- Badging individual shortcuts — not supported by the API

## References

- [MDN: shortcuts manifest member](https://developer.mozilla.org/en-US/docs/Web/Manifest/shortcuts)
- [Chrome: Define app shortcuts](https://developer.chrome.com/docs/capabilities/shortcuts)
