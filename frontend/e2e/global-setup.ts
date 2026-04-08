import { execSync } from 'child_process';
import { existsSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));

/**
 * Global setup runs once before all tests.
 * Builds the e2e server binary if it doesn't already exist (CI pre-builds it).
 */
export default function globalSetup() {
  const binaryPath = join(__dirname, 'fixtures/bin/chatto');
  if (existsSync(binaryPath)) {
    console.log('E2E server binary already exists, skipping build.');
    return;
  }

  console.log('Building e2e server...');
  execSync('mise build-e2e-server', {
    stdio: 'inherit',
    cwd: process.cwd()
  });
}
