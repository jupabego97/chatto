---
# chatto-t7hd
title: Mention Autocomplete
status: draft
type: feature
priority: normal
created_at: 2026-01-20T11:51:30Z
updated_at: 2026-02-08T13:31:43Z
---

Show a popup user picker when typing `@` in the message input, with keyboard navigation and tab completion.

## Overview

Currently, users must type exact usernames when mentioning someone. This feature adds an autocomplete popup that appears when the user types `@`, showing matching space members and allowing selection via mouse or keyboard.

## User Experience

### Triggering the Popup
1. User types `@` in message input
2. Popup appears above/below cursor position
3. Shows list of space members, filtered as user types
4. Initial state: shows all members (or most recently active)

### Filtering
- As user types after `@`, list filters to matching usernames/display names
- Match on both `login` (username) and `displayName`
- Case-insensitive matching
- Fuzzy matching optional (exact prefix match for V1)

### Selection Methods
1. **Mouse click**: Click on a user to select
2. **Arrow keys**: Navigate up/down through the list
3. **Tab/Enter**: Select the highlighted item
4. **Escape**: Close popup without selecting

### After Selection
- Replace `@partial` with `@username` (using login, not display name)
- Add a space after the username
- Close the popup
- Focus returns to message input at cursor position

### Visual Design
```
┌─────────────────────────────┐
│ @ali                        │  ← Input field
├─────────────────────────────┤
│ ┌─────────────────────────┐ │
│ │ 🟢 alice (Alice Smith)  │ │  ← Highlighted (keyboard focus)
│ │    alicia (Alicia Keys) │ │
│ │    ali_dev (Ali B.)     │ │
│ └─────────────────────────┘ │
└─────────────────────────────┘
```

### List Item Content
- User avatar (small, 24px)
- Online status indicator (green dot if online)
- Username (login) - primary text
- Display name in parentheses if different from login
- Highlight matching portion of text

### Edge Cases
- No matches: Show "No users found" message
- Many matches: Limit to 10-15, show "Type more to filter"
- Long usernames: Truncate with ellipsis
- Popup positioning: Flip above if near bottom of viewport

## Technical Design

### Data Source

Query space members on popup open:

```graphql
query GetSpaceMembersForMention($spaceId: ID!, $query: String) {
  space(id: $spaceId) {
    members(first: 15, query: $query) {
      user {
        id
        login
        displayName
        avatarUrl
        presenceStatus
      }
    }
  }
}
```

**Caching strategy:**
- Cache member list per space (invalidate on member join/leave events)
- Filter client-side for fast response
- Refetch if cache is stale (>5 min)

### Component Architecture

```
ChatInput.svelte
├── MentionAutocomplete.svelte
│   ├── MentionPopup.svelte
│   │   └── MentionItem.svelte (×N)
│   └── useMentionAutocomplete.svelte.ts (logic hook)
```

### State Management

```typescript
// useMentionAutocomplete.svelte.ts
interface MentionAutocompleteState {
  isOpen: boolean;
  query: string;           // Text after @
  triggerIndex: number;    // Position of @ in input
  selectedIndex: number;   // Keyboard navigation
  members: SpaceMember[];  // Filtered list
}

function useMentionAutocomplete(inputRef, spaceId) {
  // Returns:
  // - state (reactive)
  // - handlers: onKeyDown, onInput, onSelect, onClose
}
```

### Input Handling

Detect `@` trigger:

```typescript
function onInput(event: InputEvent) {
  const input = event.target as HTMLInputElement;
  const cursorPos = input.selectionStart ?? 0;
  const textBeforeCursor = input.value.slice(0, cursorPos);
  
  // Find last @ that isn't preceded by a word character
  const match = textBeforeCursor.match(/@(\w*)$/);
  if (match) {
    state.isOpen = true;
    state.query = match[1];
    state.triggerIndex = cursorPos - match[0].length;
  } else {
    state.isOpen = false;
  }
}
```

### Keyboard Navigation

```typescript
function onKeyDown(event: KeyboardEvent) {
  if (!state.isOpen) return;
  
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault();
      state.selectedIndex = Math.min(
        state.selectedIndex + 1,
        state.members.length - 1
      );
      break;
    case 'ArrowUp':
      event.preventDefault();
      state.selectedIndex = Math.max(state.selectedIndex - 1, 0);
      break;
    case 'Tab':
    case 'Enter':
      if (state.members.length > 0) {
        event.preventDefault();
        selectMember(state.members[state.selectedIndex]);
      }
      break;
    case 'Escape':
      event.preventDefault();
      state.isOpen = false;
      break;
  }
}
```

### Selection Logic

```typescript
function selectMember(member: SpaceMember) {
  const input = inputRef.current;
  const before = input.value.slice(0, state.triggerIndex);
  const after = input.value.slice(input.selectionStart ?? 0);
  
  // Replace @query with @username + space
  input.value = before + '@' + member.login + ' ' + after;
  
  // Move cursor after the inserted mention
  const newCursorPos = state.triggerIndex + member.login.length + 2;
  input.setSelectionRange(newCursorPos, newCursorPos);
  
  state.isOpen = false;
  input.focus();
}
```

### Popup Positioning

Calculate position relative to the @ character:

```typescript
function calculatePosition(input: HTMLElement, triggerIndex: number) {
  // Use a hidden span to measure text width up to @
  // Position popup at that x-coordinate
  // Flip above input if near viewport bottom
}
```

Or simpler approach: position popup at fixed location below/above input (less precise but easier).

### Accessibility

- `role="listbox"` on popup container
- `role="option"` on each item
- `aria-activedescendant` points to selected item
- `aria-expanded` on input indicates popup state
- `aria-autocomplete="list"` on input
- Screen reader announcements for selection changes

## Backend: Member Search API

### Option A: Client-side filtering (V1)
- Fetch all space members on popup open
- Filter in JavaScript
- Pros: Instant filtering, no network latency
- Cons: Doesn't scale to large spaces (1000+ members)

### Option B: Server-side search
- Add `query` parameter to `space.members` field
- Server filters by login/displayName
- Pros: Scales to large spaces
- Cons: Network latency on each keystroke (debounce needed)

**Recommendation:** Start with Option A for V1, add Option B when needed.

### GraphQL Schema Addition (if Option B)

```graphql
type Space {
  members(
    first: Int
    after: String
    query: String  # New: filter by login/displayName
  ): SpaceMemberConnection!
}
```

## Implementation Tasks

### Frontend Core
- [ ] Create `useMentionAutocomplete` hook with state management
- [ ] Implement @ trigger detection in input handler
- [ ] Implement keyboard navigation (arrows, tab, enter, escape)
- [ ] Create `MentionPopup.svelte` component
- [ ] Create `MentionItem.svelte` component
- [ ] Style popup (positioning, shadows, borders)
- [ ] Handle popup positioning near viewport edges

### Data Fetching
- [ ] Add space members query for autocomplete
- [ ] Implement client-side member caching
- [ ] Add filtering logic (login + displayName matching)
- [ ] Handle loading state (spinner in popup)

### Integration
- [ ] Integrate autocomplete into ChatInput component
- [ ] Wire up selection to insert @username
- [ ] Test with different input scenarios
- [ ] Handle edge cases (empty space, single member)

### Accessibility
- [ ] Add ARIA attributes to popup and items
- [ ] Test with screen reader
- [ ] Add keyboard focus indicator styling

### Testing
- [ ] Unit tests for trigger detection
- [ ] Unit tests for filtering logic
- [ ] Unit tests for selection/insertion
- [ ] E2E test: type @, select user, verify insertion

### Future: Group Mentions (Integration with chatto-jlvr)
- [ ] Add @everyone, @here, @channel to suggestion list
- [ ] Add role names to suggestion list
- [ ] Filter based on permissions (hide @everyone if no permission)
- [ ] Visual distinction for group mentions (icon, color)

## Design Decisions

- **Trigger character**: `@` only, not `:` for emoji (separate feature)
- **Minimum query length**: 0 (show all on just `@`)
- **Maximum suggestions**: 10-15 items
- **Debounce**: Not needed for client-side filtering; 200ms for server-side
- **Match algorithm**: Prefix match on login and displayName (not fuzzy)

## Dependencies

- Mentions feature (chatto-obm2) ✅ Complete
- Space member list API (existing)
- Presence system (for online indicators)

## Future Enhancements

- Fuzzy matching (typo tolerance)
- Recent mentions (show recently mentioned users first)
- Emoji autocomplete with `:` trigger
- Channel autocomplete with `#` trigger
- Slash commands with `/` trigger
