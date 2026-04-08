import { expect, type Page } from '@playwright/test';

// Shorter timeouts locally for faster feedback
const DEFAULT_POLL_TIMEOUT = process.env.CI ? 10_000 : 3_000;

/**
 * Execute a GraphQL query from within the page context.
 * Uses the page's cookies for authentication.
 */
export async function graphqlQuery<T>(
  page: Page,
  query: string,
  variables?: Record<string, unknown>
): Promise<T> {
  return page.evaluate(
    async ({ query, variables }) => {
      const response = await fetch('/api/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ query, variables })
      });
      const json = await response.json();
      if (json.errors) throw new Error(JSON.stringify(json.errors));
      return json.data;
    },
    { query, variables }
  );
}

/**
 * Wait for a space to have/not have unread rooms (server-side state).
 * Polls the server until the expected state is reached.
 */
export async function waitForSpaceUnread(
  page: Page,
  spaceId: string,
  expected: boolean,
  timeout = DEFAULT_POLL_TIMEOUT
): Promise<void> {
  await expect(async () => {
    const data = await graphqlQuery<{ space: { viewerHasUnreadRooms: boolean } }>(
      page,
      `query($id: ID!) { space(id: $id) { viewerHasUnreadRooms } }`,
      { id: spaceId }
    );
    expect(data.space.viewerHasUnreadRooms).toBe(expected);
  }).toPass({ timeout, intervals: [100, 250, 500, 1000] });
}

/**
 * Wait for a room to have/not have unread messages (server-side state).
 * Polls the server until the expected state is reached.
 */
export async function waitForRoomUnread(
  page: Page,
  spaceId: string,
  roomId: string,
  expected: boolean,
  timeout = DEFAULT_POLL_TIMEOUT
): Promise<void> {
  await expect(async () => {
    const data = await graphqlQuery<{ room: { hasUnread: boolean } }>(
      page,
      `query($spaceId: ID!, $roomId: ID!) { room(spaceId: $spaceId, roomId: $roomId) { hasUnread } }`,
      { spaceId, roomId }
    );
    expect(data.room.hasUnread).toBe(expected);
  }).toPass({ timeout, intervals: [100, 250, 500, 1000] });
}

/**
 * Wait for markRoomAsRead to complete by verifying server state.
 * Use this instead of arbitrary timeouts after entering a room.
 */
export async function waitForRoomRead(
  page: Page,
  spaceId: string,
  roomId: string,
  timeout = DEFAULT_POLL_TIMEOUT
): Promise<void> {
  await waitForRoomUnread(page, spaceId, roomId, false, timeout);
}

/**
 * Wait for a user to be deleted (no longer exists in the system).
 * Polls the server until the user query returns null or throws a "not found" error.
 */
export async function waitForUserDeleted(
  page: Page,
  userId: string,
  timeout = DEFAULT_POLL_TIMEOUT
): Promise<void> {
  await expect(async () => {
    const result = await page.evaluate(
      async ({ query, variables }) => {
        const response = await fetch('/api/graphql', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ query, variables })
        });
        const json = await response.json();
        // User is deleted if: errors exist (not found) OR user field is null
        if (json.errors || json.data?.user === null) {
          return { deleted: true };
        }
        return { deleted: false, user: json.data?.user };
      },
      {
        query: `query($id: ID!) { user(id: $id) { id } }`,
        variables: { id: userId }
      }
    );
    expect(result.deleted).toBe(true);
  }).toPass({ timeout, intervals: [100, 250, 500, 1000] });
}

/**
 * Wait for a space's member count to reach the expected value.
 * Useful for verifying membership changes after user deletion.
 */
export async function waitForSpaceMemberCount(
  page: Page,
  spaceId: string,
  expectedCount: number,
  timeout = DEFAULT_POLL_TIMEOUT
): Promise<void> {
  await expect(async () => {
    const data = await graphqlQuery<{ space: { memberCount: number } }>(
      page,
      `query($id: ID!) { space(id: $id) { memberCount } }`,
      { id: spaceId }
    );
    expect(data.space.memberCount).toBe(expectedCount);
  }).toPass({ timeout, intervals: [100, 250, 500, 1000] });
}

/**
 * Get the room ID for a room by name in a space.
 * Useful when tests need to reference rooms by ID for GraphQL queries.
 */
export async function getRoomIdByName(
  page: Page,
  spaceId: string,
  roomName: string
): Promise<string> {
  const data = await graphqlQuery<{
    me: { rooms: Array<{ id: string; name: string }> };
  }>(
    page,
    `query($spaceId: ID!) {
			me {
				rooms(spaceId: $spaceId) {
					id
					name
				}
			}
		}`,
    { spaceId }
  );

  const room = data.me.rooms.find((r) => r.name === roomName);
  if (!room) {
    throw new Error(`Room "${roomName}" not found in space ${spaceId}`);
  }
  return room.id;
}
