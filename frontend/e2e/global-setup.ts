import { execSync } from 'child_process';
import { existsSync } from 'fs';
import { join } from 'path';

/**
 * Global setup runs once before all tests. Always invokes
 * `mise build-e2e-server` — mise's source/output tracking turns this
 * into a no-op when nothing has changed and a real rebuild when backend
 * code has, so iterating on backend + e2e together doesn't silently use
 * a stale binary.
 */
export default function globalSetup() {
  if (process.env.CHATTO_E2E_SKIP_GLOBAL_BUILD === '1') {
    const binaryPath = join(process.cwd(), 'e2e/fixtures/bin/chatto');
    if (!existsSync(binaryPath)) {
      throw new Error(
        `CHATTO_E2E_SKIP_GLOBAL_BUILD is set, but the E2E server binary is missing at ${binaryPath}`
      );
    }
    return;
  }

  execSync('mise build-e2e-server', { stdio: 'inherit', cwd: process.cwd() });
}
