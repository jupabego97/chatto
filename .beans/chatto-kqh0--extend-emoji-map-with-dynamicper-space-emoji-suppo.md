---
# chatto-kqh0
title: Extend emoji map with dynamic/per-space emoji support
status: draft
type: feature
priority: normal
created_at: 2025-12-14T11:34:04Z
updated_at: 2026-01-02T20:46:00Z
parent: chatto-03l2
---

Extend the current static emoji map to support more emoji and potentially per-space customization.

## Context

The current implementation uses a static emoji map in `cli/internal/core/emoji.go` with ~70 common emoji. This works for the quick picker but has limitations:

- Users may want emoji not in the preset list
- Spaces may want custom emoji (like Slack workspaces)
- The inline `:emoji:` syntax will need a larger vocabulary

## Options to Consider

### 1. Expand static map
- Add more emoji to the existing map
- Simple, no storage changes
- Still limited to predefined set

### 2. Dynamic emoji registry
- Store emoji definitions in a KV bucket
- Allow runtime additions
- Could seed from a comprehensive list (gemoji, emoji-data, etc.)

### 3. Per-space custom emoji
- Spaces can upload/define custom emoji
- Stored in `SPACE_{spaceId}_EMOJI` bucket
- Requires image storage for custom emoji
- Adds complexity (moderation, storage costs)

### 4. Hybrid approach
- Large built-in set (option 2)
- Per-space overrides/additions (option 3)
- Most flexible but most complex

## Questions to Answer

- What's the priority? (quick picker works fine, this is enhancement)
- Do we need custom emoji or just more Unicode emoji?
- How does this interact with future inline emoji (`:wave:` in messages)?

## Related

- Parent feature: chatto-6col (emoji reactions)
- Current implementation: `cli/internal/core/emoji.go`