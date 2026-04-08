import { defineConfig } from '@playwright/test';
import baseConfig from './playwright.config';

/**
 * Flaky E2E tests: more retries to accommodate timing-sensitive tests.
 */
export default defineConfig({
  ...baseConfig,
  retries: 5
});
