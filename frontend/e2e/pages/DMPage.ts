import { expect, type Locator, type Page } from '@playwright/test';
import * as routes from '../routes';
import { RoomPage } from './RoomPage';

/**
 * Page object for Direct Messages interactions.
 * Handles the DM list, conversation navigation, and starting DMs via GraphQL API.
 */
export class DMPage {
  constructor(readonly page: Page) {}

  // --- Locators ---

  /** The DM conversation list container */
  get conversationList(): Locator {
    return this.page.locator('[data-testid="dm-list"]');
  }

  // --- Navigation ---

  /**
   * Navigate to the DM list page.
   */
  async goto(): Promise<void> {
    await this.page.goto(routes.dm);
    await this.page.waitForURL(routes.dm);
  }

  // --- API Actions ---

  /**
   * Start a DM conversation with a user via the GraphQL API and navigate to it.
   * This bypasses the UI since DMs are started from the profile popover in the
   * member list. This helper is for test setup convenience.
   */
  async startConversation(username: string): Promise<RoomPage> {
    // Look up user by login
    const userResult = await this.page.request.post('/api/graphql', {
      headers: { 'Content-Type': 'application/json', 'X-REQUEST-TYPE': 'GraphQL' },
      data: {
        query: `query FindUserByLogin($login: String!) { userByLogin(login: $login) { id } }`,
        variables: { login: username }
      }
    });
    const userData = await userResult.json();
    const userId = userData.data?.userByLogin?.id;
    if (!userId) {
      throw new Error(`User not found: ${username}`);
    }

    // Start DM
    const dmResult = await this.page.request.post('/api/graphql', {
      headers: { 'Content-Type': 'application/json', 'X-REQUEST-TYPE': 'GraphQL' },
      data: {
        query: `mutation StartDM($input: StartDMInput!) { startDM(input: $input) { id } }`,
        variables: { input: { participantIds: [userId] } }
      }
    });
    const dmData = await dmResult.json();
    const conversationId = dmData.data?.startDM?.id;
    if (!conversationId) {
      throw new Error(`Failed to start DM with ${username}`);
    }

    // Navigate to the conversation
    await this.page.goto(routes.dmConversation(conversationId));
    await this.page.waitForURL(routes.patterns.anyDmConversation);

    const roomPage = new RoomPage(this.page);
    await expect(roomPage.messageInput).toBeVisible({ timeout: 5000 });
    return roomPage;
  }

  // --- Conversation List ---

  /**
   * Get a conversation item by the other user's display name.
   * Scoped to the sidebar nav to avoid matching the room header.
   * Uses filter with exact text matching to avoid partial matches with similar names.
   */
  getConversation(displayName: string): Locator {
    return this.page
      .locator('nav a.sidebar-item')
      .filter({ has: this.page.getByText(displayName, { exact: true }) });
  }

  /**
   * Click on a conversation to open it.
   * Returns a RoomPage for interacting with messages.
   */
  async openConversation(displayName: string): Promise<RoomPage> {
    await this.getConversation(displayName).click();
    await this.page.waitForURL(routes.patterns.anyDmConversation);
    const roomPage = new RoomPage(this.page);
    // Wait for room to be ready (message input visible)
    await expect(roomPage.messageInput).toBeVisible({ timeout: 5000 });
    return roomPage;
  }

  // --- Assertions ---

  /**
   * Assert that a conversation with the given display name is visible in the list.
   */
  async expectConversationVisible(displayName: string): Promise<void> {
    await expect(this.getConversation(displayName)).toBeVisible();
  }

  /**
   * Assert that a conversation with the given display name is NOT visible.
   */
  async expectConversationNotVisible(displayName: string): Promise<void> {
    await expect(this.getConversation(displayName)).not.toBeVisible();
  }

  /**
   * Assert that the conversation header shows the expected user name.
   */
  async expectConversationHeader(displayName: string): Promise<void> {
    await expect(this.page.getByRole('heading', { name: displayName })).toBeVisible();
  }

  /**
   * Assert that a conversation is at the top of the list (first position).
   */
  async expectConversationAtTop(displayName: string): Promise<void> {
    const firstConversation = this.page.locator('nav a.sidebar-item').first();
    await expect(firstConversation).toContainText(displayName);
  }

  /**
   * Assert that a conversation has an unread indicator.
   */
  async expectConversationUnread(displayName: string): Promise<void> {
    const conv = this.getConversation(displayName);
    await expect(conv.locator('.rounded-full.bg-warning')).toBeVisible();
  }

  /**
   * Assert that a conversation does NOT have an unread indicator.
   */
  async expectConversationRead(displayName: string): Promise<void> {
    const conv = this.getConversation(displayName);
    await expect(conv.locator('.rounded-full.bg-warning')).not.toBeVisible();
  }

  /**
   * Get the display names of all conversations in order.
   */
  async getConversationOrder(): Promise<string[]> {
    const conversations = this.page.locator('nav a.sidebar-item');
    const count = await conversations.count();
    const names: string[] = [];
    for (let i = 0; i < count; i++) {
      const text = await conversations.nth(i).locator('span.truncate').textContent();
      if (text) names.push(text.trim());
    }
    return names;
  }
}
