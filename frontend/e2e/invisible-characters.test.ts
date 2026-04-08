import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';

test.describe('Invisible character messages', () => {
  test('submit button is disabled when input contains only invisible characters', async ({
    page,
    chatPage,
    roomPage
  }) => {
    // Create user and space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Verify send button is disabled initially (empty input)
    await expect(roomPage.sendButton).toBeDisabled();

    // Type only zero-width spaces (U+200B)
    await roomPage.messageInput.fill('\u200B\u200B\u200B');

    // Send button should still be disabled
    await expect(roomPage.sendButton).toBeDisabled();

    // Try with mixed invisible characters
    await roomPage.messageInput.fill('\u200B\u200C\u200D\u2060\uFEFF');
    await expect(roomPage.sendButton).toBeDisabled();

    // Try with whitespace and invisible characters
    await roomPage.messageInput.fill('  \u200B  \t\u200C\n');
    await expect(roomPage.sendButton).toBeDisabled();
  });

  test('normal messages can still be posted', async ({ page, chatPage, roomPage }) => {
    // Create user and space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type a normal message
    const testMessage = `Test message ${Date.now()}`;
    await roomPage.messageInput.fill(testMessage);

    // Send button should be enabled
    await expect(roomPage.sendButton).toBeEnabled();

    // Send the message
    await roomPage.sendButton.click();

    // Verify the message appears
    await expect(page.getByText(testMessage)).toBeVisible();

    // Input should be cleared
    await expect(roomPage.messageInput).toHaveText('');
  });

  test('messages with invisible characters mixed with visible text can be posted', async ({
    page,
    chatPage,
    roomPage
  }) => {
    // Create user and space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type message with invisible characters mixed in
    const timestamp = Date.now();
    const visiblePart = `Hello ${timestamp}`;
    const messageWithInvisible = `\u200B${visiblePart}\u200B`;
    await roomPage.messageInput.fill(messageWithInvisible);

    // Send button should be enabled because there's visible content
    await expect(roomPage.sendButton).toBeEnabled();

    // Send the message
    await roomPage.sendButton.click();

    // Verify the visible part of the message appears
    await expect(page.getByText(visiblePart)).toBeVisible();
  });

  test('emoji-only messages can be posted', async ({ page, chatPage, roomPage }) => {
    // Create user and space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type emoji only
    await roomPage.messageInput.fill('😀🎉');

    // Send button should be enabled
    await expect(roomPage.sendButton).toBeEnabled();

    // Send the message
    await roomPage.sendButton.click();

    // Verify the emoji appears
    await expect(page.getByText('😀🎉')).toBeVisible();
  });
});
