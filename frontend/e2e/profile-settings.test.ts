import { test, expect } from './setup';
import { SettingsPage } from './pages/SettingsPage';
import { createAndLoginTestUser } from './fixtures/testUser';

test.describe('Profile Settings - Display Name Validation', () => {
  test('can update display name with valid ASCII characters', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Update to a new valid name
    await settingsPage.updateDisplayName('John Doe');

    // Verify the input has the new value
    await settingsPage.expectDisplayNameValue('John Doe');
  });

  test('can update display name with international characters', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Update to Japanese name
    await settingsPage.updateDisplayName('田中太郎');
    await settingsPage.expectDisplayNameValue('田中太郎');
  });

  test('can update display name with emoji', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Update to name with emoji
    await settingsPage.updateDisplayName('Alice 🚀');
    await settingsPage.expectDisplayNameValue('Alice 🚀');
  });

  test('can update display name with German umlaut', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    await settingsPage.updateDisplayName('Müller');
    await settingsPage.expectDisplayNameValue('Müller');
  });

  test('can update display name with allowed punctuation', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Test hyphen
    await settingsPage.updateDisplayName('Mary-Jane');
    await settingsPage.expectDisplayNameValue('Mary-Jane');

    // Test apostrophe
    await settingsPage.updateDisplayName("O'Brien");
    await settingsPage.expectDisplayNameValue("O'Brien");

    // Test period
    await settingsPage.updateDisplayName('Dr. Smith');
    await settingsPage.expectDisplayNameValue('Dr. Smith');

    // Test underscore
    await settingsPage.updateDisplayName('Cool_User');
    await settingsPage.expectDisplayNameValue('Cool_User');
  });

  test('rejects display name exceeding 32 characters', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // 33 characters should be rejected
    await settingsPage.submitDisplayName('A'.repeat(33));
    await settingsPage.expectErrorVisible('cannot exceed 32 characters');

    // Should not show success message
    await expect(page.getByText('Profile updated')).not.toBeVisible();
  });

  test('accepts display name at exactly 32 characters', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    const name32 = 'A'.repeat(32);
    await settingsPage.updateDisplayName(name32);
    await settingsPage.expectDisplayNameValue(name32);
  });

  test('rejects display name with consecutive spaces', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Try to submit with consecutive spaces
    await settingsPage.submitDisplayName('John  Doe');

    // Should show error about consecutive spaces
    await settingsPage.expectErrorVisible('consecutive spaces');

    // Should not show success message
    await expect(page.getByText('Profile updated')).not.toBeVisible();
  });

  test('rejects display name with disallowed punctuation', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Test @ sign
    await settingsPage.submitDisplayName('user@domain');
    await settingsPage.expectErrorVisible('can only contain');

    // Test semicolon
    await settingsPage.submitDisplayName('John; DROP TABLE');
    await settingsPage.expectErrorVisible('can only contain');

    // Test exclamation mark
    await settingsPage.submitDisplayName('Hello!');
    await settingsPage.expectErrorVisible('can only contain');
  });

  test('rejects empty display name', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Clear the input and try to submit
    await settingsPage.displayNameInput.clear();
    await settingsPage.saveDisplayNameButton.click();

    // Should show error about empty name
    await settingsPage.expectErrorVisible('cannot be empty');
  });

  test('display name update persists across page reload', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Update display name
    const newName = `Updated Name ${Date.now()}`;
    await settingsPage.updateDisplayName(newName);

    // Reload the page
    await page.reload();

    // Verify the name persisted
    await settingsPage.expectDisplayNameValue(newName);
  });

  test('display name with mixed scripts is accepted', async ({ page }) => {
    await createAndLoginTestUser(page);
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();

    // Mixed Latin and CJK
    await settingsPage.updateDisplayName('John 田中');
    await settingsPage.expectDisplayNameValue('John 田中');
  });
});
