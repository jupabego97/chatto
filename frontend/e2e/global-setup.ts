import { execSync } from 'child_process';

/**
 * Global setup runs once before all tests. Usually invokes
 * `mise build-e2e-server` — mise's source/output tracking turns this
 * into a no-op when nothing has changed and a real rebuild when backend
 * code has, so iterating on backend + e2e together doesn't silently use
 * a stale binary. CI can opt out after building the binary explicitly inside
 * the E2E runner container.
 */
export default function globalSetup() {
  if (process.env.CHATTO_E2E_SKIP_GLOBAL_BUILD === '1') {
    return;
  }

  execSync('mise build-e2e-server', { stdio: 'inherit', cwd: process.cwd() });
}
