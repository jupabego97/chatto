import { expect } from '@playwright/test';
import { test } from './setup';
import * as routes from './routes';
import { TIMEOUTS } from './constants';
import { adminGraphql, createBootstrapAdminRequest } from './fixtures/adminRequest';
import { graphqlQuery } from './fixtures/graphqlHelpers';
import { createAndLoginTestUser } from './fixtures/testUser';

async function createJoinedRoom(page: import('@playwright/test').Page, name: string) {
  const adminContext = await createBootstrapAdminRequest(new URL(page.url()).origin);
  try {
    const groupData = await adminGraphql<{ server: { roomGroups: { id: string }[] } }>(
      adminContext,
      `query { server { roomGroups { id } } }`
    );
    const groupId = groupData.server.roomGroups[0]?.id;
    if (!groupId) throw new Error('No room group available for e2e room creation');

    const createData = await adminGraphql<{ createRoom: { id: string; name: string } }>(
      adminContext,
      `mutation($input: CreateRoomInput!) { createRoom(input: $input) { id name } }`,
      { input: { name, groupId } }
    );

    await graphqlQuery<{ joinRoom: { id: string } }>(
      page,
      `mutation($input: JoinRoomInput!) { joinRoom(input: $input) { id } }`,
      { input: { roomId: createData.createRoom.id } }
    );

    return createData.createRoom;
  } finally {
    await adminContext.dispose();
  }
}

async function renameRoom(page: import('@playwright/test').Page, roomId: string, name: string) {
  const adminContext = await createBootstrapAdminRequest(new URL(page.url()).origin);
  try {
    await adminGraphql<{ updateRoom: { id: string; name: string } }>(
      adminContext,
      `mutation($input: UpdateRoomInput!) { updateRoom(input: $input) { id name } }`,
      { input: { roomId, name } }
    );
  } finally {
    await adminContext.dispose();
  }
}

test('room URLs use names and canonicalize legacy or stale segments', async ({ page }) => {
  await createAndLoginTestUser(page);
  await page.goto(routes.chat);

  const stamp = Date.now().toString(36);
  const initialName = `urlroom${stamp}`;
  const currentName = `urlroomcurrent${stamp}`;
  const room = await createJoinedRoom(page, initialName);

  await page.goto(routes.room(room.id));
  await page.waitForURL((url) => url.pathname === routes.room(initialName), {
    timeout: TIMEOUTS.REALTIME_EVENT
  });
  await expect(page.getByRole('heading', { name: `# ${initialName}` })).toBeVisible();

  await renameRoom(page, room.id, currentName);

  await page.goto(routes.room(initialName));
  await page.waitForURL((url) => url.pathname === routes.room(currentName), {
    timeout: TIMEOUTS.REALTIME_EVENT
  });
  await expect(page.getByRole('heading', { name: `# ${currentName}` })).toBeVisible();

  await page.goto(routes.room(`missing${stamp}`));
  await page.waitForURL((url) => url.pathname === routes.chat, {
    timeout: TIMEOUTS.REALTIME_EVENT
  });
});
