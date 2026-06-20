import { describe, expect, it } from 'vitest';
import { isServerOutdatedForClient, parseSemver, serverVersionWarning } from './versionCompatibility';

describe('version compatibility', () => {
	it('parses semver with optional v-prefix and prerelease metadata', () => {
		expect(parseSemver('v0.3.7-alpha.1')).toEqual({ major: 0, minor: 3, patch: 7 });
	});

	it('does not warn for patch-only differences', () => {
		expect(isServerOutdatedForClient('0.3.6', '0.3.7')).toBe(false);
	});

	it('warns when the server minor version is older than the client', () => {
		expect(isServerOutdatedForClient('0.2.9', '0.3.0')).toBe(true);
	});

	it('warns when the server major version is older than the client', () => {
		expect(isServerOutdatedForClient('0.9.9', '1.0.0')).toBe(true);
	});

	it('does not warn when versions cannot be compared', () => {
		expect(isServerOutdatedForClient('0.2.0', 'dev')).toBe(false);
		expect(isServerOutdatedForClient(null, '0.3.0')).toBe(false);
	});

	it('returns copy for outdated servers', () => {
		expect(serverVersionWarning('0.2.9', '0.3.0')).toContain('Some features may not work');
		expect(serverVersionWarning('0.3.6', '0.3.7')).toBeNull();
	});
});
