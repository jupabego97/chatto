import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';
import { TIMEOUTS } from './constants';

/**
 * Locator helper for the emoji autocomplete popup.
 * The popup is a div with a <ul> containing emoji buttons with ":shortcode:" text.
 * Uses [^:]+ instead of \w+ because some emoji names contain non-word chars (e.g., "+1").
 */
function getPopup(page: import('@playwright/test').Page) {
  return page.locator('ul').filter({
    has: page.locator('button', { hasText: /^.+ :[^:]+:$/ })
  });
}

test.describe('Emoji autocomplete', () => {
  test('popup appears when typing colon followed by at least 2 characters', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type :he — should trigger popup (matches "heart", "heavy_heart_exclamation", etc.)
    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':he');

    // Popup should appear with emoji results
    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // Should show at least one result containing ":heart:"
    await expect(popup.locator('button', { hasText: ':heart:' })).toBeVisible();
  });

  test('popup does NOT appear with fewer than 2 characters after colon', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type just ":" — popup should not appear
    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':');
    const popup = getPopup(page);
    await expect(popup).not.toBeVisible();

    // Type one more char ":h" — still not enough (need 2)
    await roomPage.messageInput.pressSequentially('h');
    await expect(popup).not.toBeVisible();

    // Type one more ":he" — NOW popup should appear
    await roomPage.messageInput.pressSequentially('e');
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });
  });

  test('selecting emoji via Enter inserts it and closes popup', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':heart');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // The first result should be "heart" (exact match scores highest)
    const firstResult = popup.locator('button').first();
    await expect(firstResult).toContainText(':heart:');

    // Press Enter to select
    await roomPage.messageInput.press('Enter');

    // Popup should close
    await expect(popup).not.toBeVisible();

    // Input should contain the emoji character followed by a space (not ":heart")
    const value = await roomPage.messageInput.textContent();
    expect(value).toMatch(/^❤️ $/);
  });

  test('selecting emoji via click inserts it and closes popup', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':thumbsup');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // Click the "+1" / "thumbsup" emoji result
    const thumbsupButton = popup.locator('button', { hasText: ':+1:' });
    await expect(thumbsupButton).toBeVisible();
    await thumbsupButton.click();

    // Popup should close
    await expect(popup).not.toBeVisible();

    // Input should contain the thumbsup emoji + space
    const value = await roomPage.messageInput.textContent();
    expect(value).toContain('👍');
    expect(value).toMatch(/👍 $/);
  });

  test('arrow keys navigate the popup selection', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':sm');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // First item should be highlighted (has menu-item-active class)
    const buttons = popup.locator('button');
    const count = await buttons.count();
    expect(count).toBeGreaterThan(1);

    // First button should have highlighted class
    await expect(buttons.first()).toHaveClass(/menu-item-active/);

    // Press ArrowDown - second item should now be highlighted
    await roomPage.messageInput.press('ArrowDown');
    await expect(buttons.nth(1)).toHaveClass(/menu-item-active/);
    await expect(buttons.first()).not.toHaveClass(/menu-item-active/);

    // Press ArrowUp - first item should be highlighted again
    await roomPage.messageInput.press('ArrowUp');
    await expect(buttons.first()).toHaveClass(/menu-item-active/);
  });

  test('Escape closes popup without inserting emoji', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':heart');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // Press Escape
    await roomPage.messageInput.press('Escape');

    // Popup should close
    await expect(popup).not.toBeVisible();

    // Input should still contain the original text (no emoji inserted)
    await expect(roomPage.messageInput).toHaveText(':heart');
  });

  test('selected emoji is sent and rendered in the message', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type text, then emoji shortcode
    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially('Hello :wave');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // Select the first result (should be "wave" — 👋)
    await roomPage.messageInput.press('Enter');
    await expect(popup).not.toBeVisible();

    // Input should now be "Hello 👋 " — send it
    const value = await roomPage.messageInput.textContent();
    expect(value).toContain('👋');

    // Press Enter to send the message
    await roomPage.messageInput.press('Enter');

    // Message should appear with the emoji rendered
    await expect(page.locator('[role="article"]', { hasText: '👋' })).toBeVisible({
      timeout: TIMEOUTS.UI_STANDARD
    });
  });

  test('emoji autocomplete works mid-message', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type "I feel " then ":jo" — should match "joy" emoji
    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially('I feel :jo');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });
    await expect(popup.locator('button', { hasText: ':joy:' })).toBeVisible();

    // Select joy emoji
    await roomPage.messageInput.press('Enter');
    await expect(popup).not.toBeVisible();

    // Input should have "I feel 😂 " (joy = 😂)
    const value = await roomPage.messageInput.textContent();
    expect(value).toMatch(/^I feel 😂 $/);

    // Continue typing after the emoji
    await roomPage.messageInput.pressSequentially('today!');
    await expect(roomPage.messageInput).toHaveText('I feel 😂 today!');
  });

  test('popup disappears when deleting characters below threshold', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':he');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // Delete one character — now only ":h" (1 char, below threshold)
    await roomPage.messageInput.press('Backspace');
    await expect(popup).not.toBeVisible();
  });

  test('Tab selects emoji from popup (same as Enter)', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':fire');

    const popup = getPopup(page);
    await expect(popup).toBeVisible({ timeout: TIMEOUTS.UI_FAST });

    // Press Tab to select (should work like Enter)
    await roomPage.messageInput.press('Tab');

    // Popup should close
    await expect(popup).not.toBeVisible();

    // Input should contain the fire emoji
    const value = await roomPage.messageInput.textContent();
    expect(value).toContain('🔥');
  });

  test('popup does not appear for non-matching queries', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Type something that won't match any emoji
    await roomPage.messageInput.click();
    await roomPage.messageInput.pressSequentially(':zzzzqqqq');

    const popup = getPopup(page);
    await expect(popup).not.toBeVisible();
  });
});
