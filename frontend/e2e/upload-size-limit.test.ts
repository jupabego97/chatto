import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';
import { TIMEOUTS } from './constants';

// Use a very low max upload size (1 KB) so we can test with tiny files
test.use({
  serverOptions: {
    env: {
      CHATTO_CORE_ASSETS_MAX_UPLOAD_SIZE: '1KB'
    }
  }
});

test.describe('upload size limit', () => {
  test('shows error toast when file exceeds max upload size', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // The test fixture image (brighton.jpg) is much larger than 1 KB,
    // so it should be rejected by client-side validation
    await roomPage.fileInput.setInputFiles('e2e/fixtures/brighton.jpg');

    // Should see an error toast about file being too large
    await expect(page.getByText('too large')).toBeVisible({ timeout: TIMEOUTS.UI_STANDARD });

    // Attachment preview should NOT appear (file was rejected)
    await expect(roomPage.attachmentPreview).not.toBeVisible();
  });

  test('drop zone rejects oversized files with toast', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Drop a file that exceeds the 1 KB limit
    await roomPage.simulateFileDrop('e2e/fixtures/brighton.jpg');

    // Should see an error toast
    await expect(page.getByText('too large')).toBeVisible({ timeout: TIMEOUTS.UI_STANDARD });

    // Attachment preview should NOT appear
    await expect(roomPage.attachmentPreview).not.toBeVisible();
  });
});
