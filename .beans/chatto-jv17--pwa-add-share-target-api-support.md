---
# chatto-jv17
title: 'PWA: Add Share Target API support'
status: draft
type: feature
created_at: 2026-02-07T19:41:27Z
updated_at: 2026-02-07T19:41:27Z
---

## Summary

Add Web Share Target API support to Chatto's PWA manifest so users can share text, URLs, and files from other apps (gallery, browser, etc.) directly into a Chatto conversation.

This is especially valuable on mobile where the OS share sheet is a primary interaction pattern. Sharing a photo from the camera roll or a link from a browser into a chat room should be seamless.

## Background

The [Share Target API](https://developer.chrome.com/docs/capabilities/web-apis/web-share-target) is a manifest-only declaration (plus a receiving page) that registers the PWA as a share target in the OS share sheet. When a user selects Chatto from the share sheet, the browser navigates to a designated URL with the shared data as query params (for text) or as a multipart POST body (for files).

**Browser support:** Chromium-based browsers (Chrome, Edge, Samsung Internet) and Safari 15.4+. Firefox does not support it yet.

## Requirements

### Manifest Changes

Add `share_target` to `manifest.webmanifest`:

```json
{
  "share_target": {
    "action": "/chat/share",
    "method": "POST",
    "enctype": "multipart/form-data",
    "params": {
      "title": "title",
      "text": "text",
      "url": "url",
      "files": [
        {
          "name": "media",
          "accept": ["image/*", "video/*", "audio/*", "application/pdf", "text/*"]
        }
      ]
    }
  }
}
```

### Receiving Page (`/chat/share`)

Create a new SvelteKit route at `/chat/share` that:

- [ ] Handles the incoming share data (text, URL, files from the POST body or query params)
- [ ] If the user is not authenticated, redirect to login with a return URL back to the share page
- [ ] Present a "share to" picker showing the user's recent rooms and DM conversations
- [ ] Pre-fill the message composer with the shared text/URL
- [ ] For files, upload them as attachments to the selected room
- [ ] After sending, navigate to the target room

### UX Flow

1. User shares content from another app → OS share sheet shows "Chatto"
2. Chatto opens to `/chat/share` with the shared data
3. User picks a room/DM conversation from a list (sorted by recent activity)
4. Shared text/URL is pre-filled in a composer; files shown as pending attachments
5. User taps Send → message posted to selected room → navigates to that room

### Edge Cases

- [ ] User is not logged in → redirect to login, preserve share data, resume after auth
- [ ] Large files → respect existing upload limits, show error if exceeded
- [ ] Multiple files → support batch sharing
- [ ] Text-only share (no files) → pre-fill message input, no upload needed
- [ ] URL-only share → paste URL into message body (could later add link preview)

## Non-Goals

- Link previews / unfurling (separate feature)
- Web Share API (outbound sharing from Chatto) — separate concern

## References

- [MDN: Web Share Target API](https://developer.mozilla.org/en-US/docs/Web/Manifest/share_target)
- [Chrome: Receiving shared data](https://developer.chrome.com/docs/capabilities/web-apis/web-share-target)
