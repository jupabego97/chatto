import { version as clientVersion } from '$app/environment';

export interface SemverParts {
	major: number;
	minor: number;
	patch: number;
}

export function parseSemver(value: string | null | undefined): SemverParts | null {
	if (!value) return null;
	const match = value.trim().match(/^v?(\d+)\.(\d+)\.(\d+)(?:[-+].*)?$/);
	if (!match) return null;
	return {
		major: Number(match[1]),
		minor: Number(match[2]),
		patch: Number(match[3])
	};
}

export function isServerOutdatedForClient(
	serverVersion: string | null | undefined,
	currentClientVersion = clientVersion
): boolean {
	const server = parseSemver(serverVersion);
	const client = parseSemver(currentClientVersion);
	if (!server || !client) return false;

	if (server.major !== client.major) return server.major < client.major;
	return server.minor < client.minor;
}

export function serverVersionWarning(
	serverVersion: string | null | undefined,
	currentClientVersion = clientVersion
): string | null {
	if (!isServerOutdatedForClient(serverVersion, currentClientVersion)) return null;
	return `This server is running Chatto v${serverVersion}, but this client is v${currentClientVersion}. Some features may not work until the server is updated.`;
}
