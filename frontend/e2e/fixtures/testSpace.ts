import type { Page } from '@playwright/test';
import * as routes from '../routes';

export interface TestSpace {
  name: string;
  description?: string;
}

/**
 * Creates a new space via the UI.
 * Assumes the user is already logged in and on /chat.
 * Returns the space info and waits for navigation to complete.
 */
export async function createTestSpace(
  page: Page,
  options?: { name?: string; description?: string }
): Promise<TestSpace> {
  const timestamp = Date.now();
  const testSpace: TestSpace = {
    name: options?.name ?? `Test Space ${timestamp}`,
    description: options?.description
  };

  await page.getByRole('button', { name: 'Create Space' }).click();
  await page.getByLabel('Name').fill(testSpace.name);

  if (testSpace.description) {
    await page.getByLabel('Description').fill(testSpace.description);
  }

  await page.locator('button[type="submit"]').click();
  // After creating a space, user is redirected to first room: /chat/-/[spaceId]/[roomId]
  await page.waitForURL(routes.patterns.spaceOrRoom);

  return testSpace;
}
