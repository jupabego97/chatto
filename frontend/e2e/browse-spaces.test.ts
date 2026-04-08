import { test } from './setup';
import { createAndLoginTestUser } from './fixtures/testUser';
import { ExplorePage } from './pages/ExplorePage';
import { AuthPage } from './pages/AuthPage';
import * as routes from './routes';
import { TIMEOUTS } from './constants';

test.describe('Browse Spaces Directory', () => {
  test('shows joined spaces with Joined badge', async ({ page, chatPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();

    // Create a space (user becomes a member)
    await chatPage.createSpace('My Joined Space');

    // Navigate to browse spaces
    const explorePage = new ExplorePage(page);
    await explorePage.goto();

    // The joined space should appear with a "Joined" badge
    await explorePage.expectSpaceJoined('My Joined Space');
  });

  test('shows non-joined spaces with Join button', async ({ page, chatPage, browser, serverURL }) => {
    // First user creates a space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace('Public Space');

    // Second user browses spaces
    const context = await browser.newContext({ baseURL: serverURL });
    const otherPage = await context.newPage();
    await createAndLoginTestUser(otherPage);

    const explorePage = new ExplorePage(otherPage);
    await explorePage.goto();

    // The space should show a Join button (not a Joined badge)
    await explorePage.expectSpaceJoinable('Public Space');

    await context.close();
  });

  test('clicking a joined space card navigates to the space', async ({ page, chatPage }) => {
    await createAndLoginTestUser(page);
    await chatPage.goto();

    // Create a space
    await chatPage.createSpace('Navigate Test Space');

    // Navigate to browse spaces
    const explorePage = new ExplorePage(page);
    await explorePage.goto();

    // Click the Joined badge link on the space card
    const spaceCard = explorePage.getSpaceItem('Navigate Test Space');
    await spaceCard.getByRole('link', { name: 'Joined' }).click();

    // Should navigate to the space (not stay on /chat/spaces)
    await page.waitForURL(routes.patterns.anySpace);
  });

  test('spaces load after new user registers via UI', async ({
    page,
    chatPage,
    browser,
    serverURL
  }) => {
    // First user creates a space so there's something to browse
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace('Visible After Registration');

    // Second user registers via the UI flow (which redirects to /chat/spaces)
    const context = await browser.newContext({ baseURL: serverURL });
    const newUserPage = await context.newPage();
    const authPage = new AuthPage(newUserPage);
    const ts = Date.now();
    await authPage.register(`newuser${ts}`, `newuser${ts}@example.com`, 'testpassword123');

    // New user has no history, so redirect chain ends at Browse Spaces
    await newUserPage.waitForURL(routes.spaces, { timeout: TIMEOUTS.POLLING_EXTENDED });

    // After registration redirect, spaces should load — not stuck on skeletons
    const explorePage = new ExplorePage(newUserPage);
    await explorePage.expectSpaceVisible('Visible After Registration');

    await context.close();
  });

  test('joining a space from browse directory navigates to it', async ({
    page,
    chatPage,
    browser,
    serverURL
  }) => {
    // First user creates a space
    await createAndLoginTestUser(page);
    await chatPage.goto();
    await chatPage.createSpace('Joinable Space');

    // Second user joins via browse directory
    const context = await browser.newContext({ baseURL: serverURL });
    const otherPage = await context.newPage();
    await createAndLoginTestUser(otherPage);

    const explorePage = new ExplorePage(otherPage);
    await explorePage.goto();
    await explorePage.joinSpace('Joinable Space');

    // Should be navigated to the space (not stay on /chat/spaces)
    await otherPage.waitForURL(routes.patterns.anySpace);

    await context.close();
  });
});
