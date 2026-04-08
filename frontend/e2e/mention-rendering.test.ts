import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';
import { ChatPage, RoomPage, ExplorePage } from './pages';
import { TIMEOUTS } from './constants';

test.describe('Mention rendering', () => {
  test('valid @mention renders as bold text', async ({ page, chatPage, roomPage }) => {
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message mentioning the current user
    await roomPage.sendMessage(`Hello @${user.login}!`);

    // The mention should be wrapped in a strong tag with "mention" class
    const mentionElement = page.locator('span.mention', { hasText: `@${user.login}` });
    await expect(mentionElement).toBeVisible();
  });

  test('invalid @mention does not render as bold', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message with an invalid mention (user doesn't exist)
    await roomPage.sendMessage('Hello @nonexistentuser!');

    // The text should be visible but NOT wrapped in a mention element
    await expect(page.getByText('@nonexistentuser')).toBeVisible();
    await expect(page.locator('span.mention', { hasText: '@nonexistentuser' })).not.toBeVisible();
  });

  test('message mentioning current user has highlight background', async ({
    page,
    chatPage,
    roomPage,
    browser,
    serverURL
  }) => {
    // User 1 creates space
    const user1 = await createAndLoginTestUser(page);
    await chatPage.goto();
    const spaceName = await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // User 2 joins and mentions user 1
    const context2 = await browser!.newContext({
      baseURL: serverURL,
      viewport: { width: 1280, height: 720 }
    });
    const page2 = await context2.newPage();

    try {
      const user2 = await createAndLoginTestUser(page2);
      const chatPage2 = new ChatPage(page2);
      const roomPage2 = new RoomPage(page2);
      const explorePage2 = new ExplorePage(page2);

      await chatPage2.goto();
      await chatPage2.goToExploreSpaces();
      await explorePage2.joinSpace(spaceName);
      await chatPage2.enterRoom('general');

      // Wait for both users to be visible
      await roomPage.expectMemberVisible(user2.login, { timeout: TIMEOUTS.UI_STANDARD });

      // User 2 sends a message mentioning user 1
      await roomPage2.sendMessage(`Hey @${user1.login}, check this out!`);

      // User 1 should see the message with highlight (bg-warning/10 class)
      const messageArticle = page
        .locator('[role="article"]')
        .filter({ hasText: `@${user1.login}` });
      await expect(messageArticle).toBeVisible({ timeout: TIMEOUTS.UI_STANDARD });
      await expect(messageArticle).toHaveClass(/bg-warning\/10/);
    } finally {
      await context2.close();
    }
  });

  test('self-authored message mentioning self does not highlight', async ({
    page,
    chatPage,
    roomPage
  }) => {
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message mentioning yourself
    await roomPage.sendMessage(`Note to myself @${user.login}`);

    // The message should NOT have the highlight background
    const messageArticle = page.locator('[role="article"]').filter({ hasText: `@${user.login}` });
    await expect(messageArticle).toBeVisible();
    await expect(messageArticle).not.toHaveClass(/bg-warning\/10/);
  });

  test('@mention in code block does not render as bold', async ({ page, chatPage, roomPage }) => {
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message with a mention inside inline code
    // Don't use sendMessage() because it waits for the raw markdown text
    await roomPage.messageInput.fill(`Use \`@${user.login}\` to mention someone`);
    await roomPage.messageInput.press('Enter');

    // Wait for the rendered code element containing the mention
    const codeElement = page.locator('code', { hasText: `@${user.login}` });
    await expect(codeElement).toBeVisible({ timeout: TIMEOUTS.UI_STANDARD });

    // There should be no span.mention inside code
    await expect(page.locator('code span.mention')).not.toBeVisible();
  });

  test('@mention in fenced code block does not render as bold', async ({
    page,
    chatPage,
    roomPage
  }) => {
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message with a fenced code block containing a mention
    await roomPage.messageInput.fill(`Check this code:\n\`\`\`\n@${user.login}\n\`\`\``);
    await roomPage.sendButton.click();

    // Wait for the mention text to appear inside the pre element (scoped to avoid
    // matching the @username in the member list sidebar)
    await expect(page.locator('pre', { hasText: `@${user.login}` })).toBeVisible();

    // The mention inside pre/code should NOT be styled
    await expect(page.locator('pre span.mention')).not.toBeVisible();
  });

  test('@mention in blockquote does not render as bold', async ({ page, chatPage, roomPage }) => {
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message with a blockquote containing a mention
    // Don't use sendMessage() because it waits for the raw markdown text
    await roomPage.messageInput.fill(`> @${user.login} said this earlier`);
    await roomPage.messageInput.press('Enter');

    // Wait for the rendered blockquote element containing the mention
    const blockquoteElement = page.locator('blockquote', { hasText: `@${user.login}` });
    await expect(blockquoteElement).toBeVisible({ timeout: TIMEOUTS.UI_STANDARD });

    // The mention inside blockquote should NOT be styled
    await expect(page.locator('blockquote span.mention')).not.toBeVisible();
  });

  test('mention is case-insensitive', async ({ page, chatPage, roomPage }) => {
    const user = await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Send a message with uppercase mention
    const upperLogin = user.login.toUpperCase();
    await roomPage.sendMessage(`Hello @${upperLogin}!`);

    // Should still render as bold mention
    const mentionElement = page.locator('span.mention', { hasText: `@${upperLogin}` });
    await expect(mentionElement).toBeVisible();
  });
});
