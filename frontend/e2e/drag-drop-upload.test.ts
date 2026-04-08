import { expect } from '@playwright/test';
import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';

test.describe('drag and drop image upload', () => {
  test('drop zone overlay appears when dragging files over room', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Initially, overlay should not be visible
    await roomPage.expectDropZoneOverlayNotVisible();

    // Simulate dragging files over the room
    await roomPage.simulateDragEnter();

    // Overlay should appear
    await roomPage.expectDropZoneOverlayVisible();
  });

  test('drop zone overlay disappears when dragging files away', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Start dragging over the room
    await roomPage.simulateDragEnter();
    await roomPage.expectDropZoneOverlayVisible();

    // Drag away from the room
    await roomPage.simulateDragLeave();

    // Overlay should disappear
    await roomPage.expectDropZoneOverlayNotVisible();
  });

  test('dropping image file adds it to attachment preview', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Drop a file onto the room
    await roomPage.dropFile('e2e/fixtures/brighton.jpg');

    // Attachment preview should be visible
    await expect(roomPage.attachmentPreview).toBeVisible();
  });

  test('can send message after dropping image', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Drop a file
    await roomPage.dropFile('e2e/fixtures/brighton.jpg');

    // Add some text and send
    const testMessage = `Drag drop test ${Date.now()}`;
    await roomPage.messageInput.fill(testMessage);
    await roomPage.messageInput.press('Enter');

    // Wait for attachment preview to clear (message sent)
    await expect(roomPage.attachmentPreview).not.toBeVisible();

    // Verify the message and attachment appear
    await roomPage.expectMessageVisible(testMessage);
    await expect(roomPage.attachmentImage).toBeVisible();
  });

  test('can drop multiple files', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Drop first file
    await roomPage.dropFile('e2e/fixtures/brighton.jpg');

    // Drop second file (use the same file, just testing multiple adds)
    await roomPage.dropFile('e2e/fixtures/brighton.jpg');

    // Should have two attachment previews
    await expect(roomPage.attachmentPreview).toHaveCount(2);
  });

  test('drop zone overlay disappears after dropping file', async ({ page, chatPage, roomPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // Start dragging
    await roomPage.simulateDragEnter();
    await roomPage.expectDropZoneOverlayVisible();

    // Drop the file
    await roomPage.simulateFileDrop('e2e/fixtures/brighton.jpg');

    // Overlay should disappear after drop
    await roomPage.expectDropZoneOverlayNotVisible();

    // But attachment preview should appear
    await expect(roomPage.attachmentPreview).toBeVisible();
  });

  test('combined: drag-drop file and click-attach file both appear', async ({
    page,
    chatPage,
    roomPage
  }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace();
    await chatPage.enterRoom('general');

    // First, select a file via the attach button
    await roomPage.selectAttachment('e2e/fixtures/brighton.jpg');
    await expect(roomPage.attachmentPreview).toHaveCount(1);

    // Then drop another file
    await roomPage.dropFile('e2e/fixtures/brighton.jpg');

    // Both should appear (2 previews)
    await expect(roomPage.attachmentPreview).toHaveCount(2);
  });
});
