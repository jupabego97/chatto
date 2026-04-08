import { test as base } from '@playwright/test';
import { startServer, stopServer, type ServerInfo, type StartServerOptions } from './fixtures/server';
import {
  AccountPage,
  AdminPage,
  AuthPage,
  ChatPage,
  RoomPage,
  ExplorePage,
  DMPage,
  NotificationsPage,
  NotificationSettingsPage,
  SpaceAdminPage,
  SpaceAdminRoomsPage,
  SpaceRolesPage
} from './pages';

// Extend the base test with our fixtures
export const test = base.extend<{
  accountPage: AccountPage;
  adminPage: AdminPage;
  authPage: AuthPage;
  chatPage: ChatPage;
  roomPage: RoomPage;
  explorePage: ExplorePage;
  dmPage: DMPage;
  notificationsPage: NotificationsPage;
  notificationSettingsPage: NotificationSettingsPage;
  spaceAdminPage: SpaceAdminPage;
  spaceAdminRoomsPage: SpaceAdminRoomsPage;
  spaceRolesPage: SpaceRolesPage;
  serverURL: string; // Expose server URL to tests for creating new contexts
  server: ServerInfo; // Test-scoped: one server per test
  serverOptions: StartServerOptions; // Override to pass custom options (e.g. env vars) to the server
}>({
  // Option fixture: tests can override via test.use({ serverOptions: { ... } })
  serverOptions: [{}, { option: true }],

  // Test-scoped: one server per test for complete isolation
  server: async ({ serverOptions }, use, testInfo) => {
    const server = await startServer(testInfo, serverOptions);
    await use(server);
    await stopServer(server, testInfo);
  },

  // Override baseURL to use the test's server
  baseURL: async ({ server }, use) => {
    await use(server.baseURL);
  },

  // Expose server URL for tests that need to create new contexts
  serverURL: async ({ server }, use) => {
    await use(server.baseURL);
  },

  // Page object fixtures
  accountPage: async ({ page }, use) => {
    await use(new AccountPage(page));
  },

  adminPage: async ({ page }, use) => {
    await use(new AdminPage(page));
  },

  authPage: async ({ page }, use) => {
    await use(new AuthPage(page));
  },

  chatPage: async ({ page }, use) => {
    await use(new ChatPage(page));
  },

  roomPage: async ({ page }, use) => {
    await use(new RoomPage(page));
  },

  explorePage: async ({ page }, use) => {
    await use(new ExplorePage(page));
  },

  dmPage: async ({ page }, use) => {
    await use(new DMPage(page));
  },

  notificationsPage: async ({ page }, use) => {
    await use(new NotificationsPage(page));
  },

  notificationSettingsPage: async ({ page }, use) => {
    await use(new NotificationSettingsPage(page));
  },

  spaceAdminPage: async ({ page }, use) => {
    await use(new SpaceAdminPage(page));
  },

  spaceAdminRoomsPage: async ({ page }, use) => {
    await use(new SpaceAdminRoomsPage(page));
  },

  spaceRolesPage: async ({ page }, use) => {
    await use(new SpaceRolesPage(page));
  }
});

// Re-export expect for convenience
export { expect } from '@playwright/test';
