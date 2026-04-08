import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';
import * as routes from './routes';

/**
 * E2E tests for the notification settings page.
 */

test.describe('Notification Settings - Page Structure', () => {
  test('notification settings page renders correctly', async ({
    page,
    notificationSettingsPage
  }) => {
    await createAndLoginTestUser(page);
    await notificationSettingsPage.goto();

    // Verify page heading is visible
    await expect(page.getByRole('heading', { name: 'Notifications', exact: true })).toBeVisible();

    // Verify notification sound section is present
    await expect(page.getByText('Notification Sound')).toBeVisible();
  });

  test('notification settings page is accessible from chat settings', async ({
    page,
    chatPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();

    // Navigate to settings
    await page.goto(routes.settings);

    // Click on Notifications in the settings sidebar (not the header notification icon)
    await page.locator('nav').getByRole('link', { name: 'Notifications', exact: true }).click();

    // Should be on the notification settings page
    await page.waitForURL(routes.settingsNotifications);
    await expect(page.getByRole('heading', { name: 'Notifications', exact: true })).toBeVisible();
  });
});

test.describe('Notification Settings - Sound Settings', () => {
  test('shows notification sound section with options', async ({
    page,
    notificationSettingsPage
  }) => {
    await createAndLoginTestUser(page);
    await notificationSettingsPage.goto();

    // Should show the notification sound section
    await expect(page.getByText('Notification Sound')).toBeVisible();

    // Should show sound categories as headings (the page shows categories like "Silent", "Simple", etc.)
    await expect(page.getByRole('heading', { name: 'Silent' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Simple' })).toBeVisible();
  });

  test('can select a different notification sound', async ({ page, notificationSettingsPage }) => {
    await createAndLoginTestUser(page);
    await notificationSettingsPage.goto();

    // Find and click on a different sound option (e.g., "Pop")
    const popSound = page.getByRole('button', { name: /Pop/ });
    await popSound.click();

    // The selected sound should have the accent styling (indicating selection)
    const popSoundRow = page.locator('button', { hasText: 'Pop' }).first();
    await expect(popSoundRow.locator('.border-accent')).toBeVisible();
  });

  test('can select silent mode', async ({ page, notificationSettingsPage }) => {
    await createAndLoginTestUser(page);
    await notificationSettingsPage.goto();

    // Find and click on the "Silent" option
    const silentOption = page.getByRole('button', { name: /Silent/ });
    await silentOption.click();

    // Verify it's selected
    const silentRow = page.locator('button', { hasText: 'Silent' }).first();
    await expect(silentRow.locator('.border-accent')).toBeVisible();
  });

  test('notification sound preference persists across page reload', async ({
    page,
    notificationSettingsPage
  }) => {
    await createAndLoginTestUser(page);
    await notificationSettingsPage.goto();

    // Select "Silent" mode
    const silentOption = page.getByRole('button', { name: /Silent/ });
    await silentOption.click();

    // Reload the page
    await page.reload();

    // Silent should still be selected
    const silentRow = page.locator('button', { hasText: 'Silent' }).first();
    await expect(silentRow.locator('.border-accent')).toBeVisible();
  });
});
