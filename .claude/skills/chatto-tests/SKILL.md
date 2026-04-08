---
name: chatto-tests
description: "Use when writing, fixing, or planning tests (unit, integration, or e2e). Covers test patterns, E2E best practices, Playwright fixtures/POMs, scroll testing with virtua, multi-user real-time tests, and email testing."
---

Think hard about the changes we've made in this conversation, branch or PR. Add any CLI and/or e2e tests that might be missing. Please also test negative cases, especially around permission checks.

## Testing

### Always Consider E2E Tests

**When implementing user-facing features or fixes, proactively suggest adding e2e tests.** Don't wait for the user to ask. If the change affects:

- User input handling (forms, chat input, file uploads)
- UI state (buttons enabled/disabled, visibility, navigation)
- Real-time behavior (messages, notifications, presence)
- Authentication or authorization flows

...then propose an e2e test as part of the implementation plan. E2e tests in `frontend/e2e/` catch integration issues that unit tests miss.

### Manually testing with the compiled application

- Compile and run the CLI using `mise run run` (not a typo)
- It will be accessible at `http://localhost:4000`.

### Unit and Integration Tests

- `mise test` to run all tests (lint + Go tests + frontend/e2e tests)
- `mise lint-frontend` to run svelte-check and eslint separately
- `mise dev` to start a full development environment. This hot-reloads both backend and frontend on code changes. Sometimes it gets stuck because it doesn't correctly shut down the embedded NATS server; if that happens, find the process that's listening on port 4555 and kill it, then try again.
- Write tests for all functionality you add or change.
- Use table-driven tests for Go wherever possible.
- Write e2e tests for user flows in the frontend using Playwright.
- Use mocks and fakes for unit tests to isolate components.
- Use Page Object Models (POMs) for E2E tests - see `frontend/e2e/pages/`.

### E2E Test Isolation

Each e2e test runs against its own isolated Chatto instance. This means:

- Tests don't share state (users, spaces, rooms, permissions)
- No cleanup is required between tests
- Tests can safely create users/spaces without worrying about collisions
- Permission changes in one test don't affect others

The test infrastructure spins up a fresh Chatto server for each test file, so you don't need to worry about restoring state after modifying permissions or other instance-level settings.

### Build Tags for CLI Tests

Always use `mise test-cli` to run Go tests, not `go test` directly. The `http_server` package requires `-tags test_endpoints` to enable mock email endpoints used by tests. Running without this tag causes test failures with 404 errors on test endpoints like `/auth/test/last-email`.

### DM Rooms Need Explicit Testing

The DM space has different initialization (system-created at startup, not user-triggered) and different membership patterns (auto-joined on creation). Changes to room, message, or space logic should always include DM-specific tests - unit tests for regular rooms passing doesn't guarantee DM rooms work.

### Run E2E Tests Before Committing Refactors

Unit tests passing doesn't guarantee the system works end-to-end. For refactoring work that touches data flow (subjects, streams, queries), run e2e tests before committing to catch integration issues.

### Permission Tests Need Both Positive and Negative Cases

When testing authorization/permissions, always test both directions:

- **Positive**: User WITH the permission CAN access/perform the action
- **Negative**: User WITHOUT the permission is DENIED access

This applies to both backend tests (resolver authorization checks) and frontend e2e tests (route guards, UI visibility). Missing negative tests means you don't know if permission checks are actually enforced.

### Permission Default Changes Cascade to E2E Tests

When modifying default role permissions (e.g., removing `room.create` from `everyone`), E2E tests break if regular members perform those actions. The tests time out rather than showing permission errors because UI elements are hidden.

**Fix pattern:**

1. Have the space owner perform privileged actions (they retain all permissions)
2. Have regular members use alternate flows (e.g., Browse Rooms -> Join instead of Create Room)

```typescript
// Before: User B (regular member) creates room - breaks if room.create removed
await chatPage2.createRoom();

// After: User A (owner) creates room, User B joins via Browse Rooms
const roomName = await chatPage.createRoom(); // User A creates
await page2.getByRole("link", { name: "Browse Rooms" }).click();
await page2
  .locator("li", { hasText: `# ${roomName}` })
  .getByRole("button", { name: "Join" })
  .click();
```

### Multi-User Real-Time E2E Tests

Real-time features require testing that events are visible to _other_ users, not just the actor. A common gap: testing that User A's action succeeds, but not that User B sees the resulting event.

**Pattern for multi-user tests:**

```typescript
test("user sees leave event when another user leaves", async ({
  page,
  browser,
  serverURL,
}) => {
  const user1 = await createAndLoginTestUser(page);

  const context2 = await browser!.newContext({ baseURL: serverURL });
  const page2 = await context2.newPage();

  try {
    const user2 = await createAndLoginTestUser(page2);

    await expect(page.getByText(`${user2.displayName} joined`)).toBeVisible({
      timeout: 5000,
    });

    await page2.getByTitle("Leave room").click();

    await expect(page.getByText(`${user2.displayName} left`)).toBeVisible({
      timeout: 5000,
    });
  } finally {
    await context2.close();
  }
});
```

**Test both directions**: If User A can trigger an event, test that User B receives it.

### API-Based Message Posting for E2E Setup

When tests need many messages for setup (e.g., making a container scrollable), use GraphQL API calls instead of UI-based posting. This is ~10x faster and more reliable:

```typescript
async function postMessagesViaAPI(
  page: Page,
  spaceId: string,
  roomId: string,
  messages: string[]
): Promise<void> {
  for (const body of messages) {
    await page.request.post('/api/graphql', {
      headers: { 'Content-Type': 'application/json', 'X-REQUEST-TYPE': 'GraphQL' },
      data: {
        query: `mutation($input: PostMessageInput!) { postMessage(input: $input) { id } }`,
        variables: { input: { spaceId, roomId, body } }
      }
    });
  }
}
```

**When to use UI-based posting**: When testing the actual posting behavior (e.g., "user posts while scrolled up" tests the scroll-to-bottom behavior triggered by sending).

### verifyRealtimeSync Limitations

The `verifyRealtimeSync` helper from `fixtures/realtimeSync.ts` only works when the receiver is at the bottom of the chat (auto-scroll enabled). If the receiver is scrolled up, the sync message won't be visible and the assertion will fail.

**Use `verifyRealtimeSync` when**: Both users are at the bottom of the chat and you need to prove WebSocket subscriptions are connected before testing.

**Use `waitForRoomReady` instead when**: The receiver might be scrolled up, or you just need to ensure the room UI is loaded. Combine with `TIMEOUTS.REALTIME_EVENT` for assertions.

### Scroll Stabilization After Bulk Posting

After posting messages via API, the UI may still be auto-scrolling. Wait for scroll position to stabilize before testing scroll behavior:

```typescript
await postMessagesViaAPI(page, spaceId, roomId, messages);
await expect(page.getByText('Message 20')).toBeVisible({ timeout: 5000 });

await expect(async () => {
  const info = await messagesContainer.evaluate((el) => ({
    scrollHeight: el.scrollHeight,
    scrollTop: el.scrollTop,
    clientHeight: el.clientHeight
  }));
  const distanceFromBottom = info.scrollHeight - info.scrollTop - info.clientHeight;
  expect(distanceFromBottom).toBeLessThan(50);
}).toPass({ timeout: 5000, intervals: [100, 250, 500, 1000] });

await scrollContainerToTop(page, messagesContainer);
```

### E2E Scrolling with virtua

The event list uses `virtua` for DOM virtualization. Programmatic `scrollTop` assignment doesn't work reliably with virtua's scroll correction mechanism. E2E tests must use **native mouse wheel events** instead:

```typescript
async function scrollContainerToTop(page: Page, container: Locator) {
  const box = await container.boundingBox();
  if (!box) throw new Error('Container not visible');
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  for (let i = 0; i < 15; i++) {
    await page.mouse.wheel(0, -800);
    await page.waitForTimeout(50);
  }
}
```

### E2E Server State Synchronization

For multi-user tests that depend on server state changes (like unread indicators), **don't use arbitrary timeouts**. Instead, poll the server via GraphQL until the expected state is reached:

```typescript
import {
  waitForSpaceUnread,
  waitForRoomUnread,
  waitForRoomRead,
  getRoomIdByName,
} from "./fixtures/graphqlHelpers";

await waitForSpaceUnread(page, spaceId, true);
await waitForRoomUnread(page, spaceId, roomId, true);
await waitForRoomRead(page, spaceId, roomId);
```

### Multi-Tab Sync Tests Need All UI Levels

When testing multi-tab/multi-device sync for indicators (like unread dots), verify ALL levels of UI sync, not just one.

```typescript
// COMPLETE - Tests both space-level AND room-level
await expect(page3.locator('[data-testid="space-unread-dot"]')).not.toBeVisible();
await expect(page3.locator('[data-testid="room-unread-dot"]')).not.toBeVisible();
```

### Playwright Fixture Naming

Playwright doesn't support underscore-prefixed fixture parameters (like `_page`). Solutions:
- Remove unused fixtures from destructuring entirely
- Use destructuring rename: `{ chatPage: _chatPage }`
- Use `eslint-disable-next-line no-empty-pattern` for empty patterns `{}`

### E2E Page Object Models

Use Page Object Models (`frontend/e2e/pages/`) to encapsulate UI interactions:

- **ChatPage**: Sidebar navigation, space creation, room entry
- **RoomPage**: Message input/sending, attachments, thread pane
- **MessageComponent**: Per-message actions (react, delete, edit, threads)
- **ExplorePage**: Space discovery and joining

POMs are injected via Playwright fixtures in `setup.ts`. For multi-user tests with a second browser context, instantiate POMs directly: `const chatPage2 = new ChatPage(page2)`.

### E2E Selector Specificity

Avoid `getByText()` for assertions - it often matches multiple elements. Prefer specific locators:

| Instead of                          | Use                                                       |
| ----------------------------------- | --------------------------------------------------------- |
| `getByText('Browse Spaces')`        | `getByRole('heading', { name: 'Browse Spaces' })`         |
| `getByText('Access Denied')`        | `getByText('Access Denied', { exact: true })`             |
| `getByText(displayName)` in sidebar | `locator('nav').getByRole('link', { name: displayName })` |

### E2E Selector Resilience

Avoid selectors that couple to specific HTML element types. Target semantic content (headings, alt text, roles) rather than structural elements.

### E2E Form Selectors with data-testid

Use `data-testid` attributes for form elements. `TextInput` and `TextArea` components accept a `testid` prop. Naming convention: `{scope}-{component}-{element}`.

### E2E Scroll Position Test Robustness

Avoid exact scroll position comparisons. Test the actual behavior:

| Flaky assertion | Robust assertion |
|-----------------|------------------|
| `expect(scrollTop).toBe(0)` | `expect(scrollTop).toBeLessThan(5)` |
| `expect(scrollTopAfter - scrollTopBefore).toBeLessThan(50)` | `expect(distanceFromBottom).toBeGreaterThan(100)` |

### E2E UI Transition Timing

When switching between views, wait for old content to disappear first, then assert new content.

### E2E JavaScript Error Monitoring

Add tests that capture console errors and page errors for real-time event handling code paths.

### Dialog Interception After `createRoom()`

After `chatPage.createRoom()`, the room creation modal may still intercept pointer events. Use `page.goto()` for navigation instead of clicking sidebar links.

### Avoid Default Room Names in E2E Tests

Don't use names like `'general'` or `'announcements'` that conflict with system-created rooms.

### Email Testing

| Tool           | Purpose                                           | Location                                   |
| -------------- | ------------------------------------------------- | ------------------------------------------ |
| `MockSender`   | Capture emails in memory for business logic tests | `internal/email/mock.go`                   |
| `go-smtp-mock` | Test actual SMTP protocol with go-mail library    | `internal/email/email_integration_test.go` |

**go-smtp-mock quirks**: Set `MultipleMessageReceiving: true`. Use `server.WaitForMessages(count, timeout)` instead of `server.Messages()`.
