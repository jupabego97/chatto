import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';
import { waitForRoomReady } from './fixtures/realtimeSync';
import { RoomPage } from './pages';
import { TIMEOUTS } from './constants';
import * as routes from './routes';

test.describe('Composer drafts', () => {
  test('drafts are tab-specific and do not leak to other tabs', async ({
    page,
    chatPage,
    roomPage,
    browser,
    serverURL
  }) => {
    // Create user and space
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Get the room URL for the second tab
    const roomUrl = page.url();

    // Type a draft message in tab 1
    const draftText = `Draft message ${Date.now()}`;
    await roomPage.messageInput.fill(draftText);

    // Verify the draft is in tab 1
    await expect(roomPage.messageInput).toHaveText(draftText);

    // Open a second tab with the same user in the same room
    const context2 = await browser!.newContext({
      baseURL: serverURL,
      viewport: { width: 1280, height: 720 }
    });
    const page2 = await context2.newPage();

    try {
      // Login as the same user in tab 2
      const loginResponse = await page2.request.post('/auth/login', {
        data: {
          login: user.login,
          password: user.password
        }
      });
      expect(loginResponse.ok()).toBeTruthy();

      // Navigate to the same room
      await page2.goto(roomUrl);
      await page2.waitForURL(routes.patterns.anyRoom);

      // The message input in tab 2 should be empty (not showing tab 1's draft)
      const roomPage2 = new RoomPage(page2);
      await expect(roomPage2.messageInput).toHaveText('');

      // Type a different draft in tab 2
      const draftText2 = `Different draft ${Date.now()}`;
      await roomPage2.messageInput.fill(draftText2);

      // Verify tab 2 has its own draft
      await expect(roomPage2.messageInput).toHaveText(draftText2);

      // Go back to tab 1 and verify its draft is unchanged
      await expect(roomPage.messageInput).toHaveText(draftText);
    } finally {
      await context2.close();
    }
  });

  test('draft persists when navigating away and back to room', async ({
    page,
    chatPage,
    roomPage
  }) => {
    // Create user and space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type a draft message
    const draftText = `Draft that should persist ${Date.now()}`;
    await roomPage.messageInput.fill(draftText);

    // Navigate to a different room
    await chatPage.enterRoom('announcements');

    // The input should be empty in the new room
    await expect(roomPage.messageInput).toHaveText('');

    // Navigate back to general
    await chatPage.enterRoom('general');

    // The draft should be restored
    await expect(roomPage.messageInput).toHaveText(draftText);
  });

  test('draft image attachments persist when navigating away and back to room', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Attach an image in general
    await roomPage.selectAttachment('e2e/fixtures/brighton.jpg');
    await expect(roomPage.attachmentPreview).toBeVisible();

    // Navigate away to announcements
    await chatPage.enterRoom('announcements');
    await expect(roomPage.attachmentPreview).not.toBeVisible();

    // Navigate back to general - attachment should be restored
    await chatPage.enterRoom('general');
    await expect(roomPage.attachmentPreview).toBeVisible();
  });
});

test.describe('Composer focus', () => {
  test('clicking empty area in composer focuses the text input', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');
    await waitForRoomReady(page, 'general');

    // Click somewhere else first to blur the composer
    await page.getByRole('heading').first().click();

    // Verify the editor is not focused
    const editor = roomPage.messageInput;
    await expect(editor).not.toBeFocused();

    // Click the composer container area (the outer padding area, not the editor itself).
    // The composer wrapper contains the input container with the editor inside.
    // Clicking its padding should focus the editor.
    const composerContainer = page.locator('.flex.flex-col.gap-2.p-2').filter({
      has: editor
    });
    const box = await composerContainer.boundingBox();
    expect(box).not.toBeNull();

    // Click near the top-left padding area of the composer (away from buttons and editor)
    await page.mouse.click(box!.x + 5, box!.y + 5);

    // The editor should now be focused
    await expect(editor).toBeFocused({ timeout: TIMEOUTS.UI_FAST });
  });

  test('clicking attach button opens file dialog, not just focus', async ({ page, chatPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');
    await waitForRoomReady(page, 'general');

    // The attach button should trigger its own behavior, not be intercepted
    const attachButton = page.getByTitle('Attach file');
    await expect(attachButton).toBeVisible();

    // Set up a listener for the file chooser dialog
    const fileChooserPromise = page.waitForEvent('filechooser', { timeout: TIMEOUTS.UI_STANDARD });
    await attachButton.click();

    // The file dialog should open (proving the button handled the click, not the composer)
    const fileChooser = await fileChooserPromise;
    expect(fileChooser).toBeTruthy();
  });
});
