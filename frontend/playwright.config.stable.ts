import { defineConfig } from '@playwright/test';
import baseConfig from './playwright.config';

/**
 * Stable E2E tests: low retry count since these should reliably pass.
 */
export default defineConfig({
  ...baseConfig,
  retries: 1
});
