/**
 * Utility for namespacing localStorage keys by instance ID.
 * Prevents data collisions when multiple Chatto instances share a browser.
 */

/**
 * Build a localStorage key scoped to a specific instance.
 *
 * @example instanceStorageKey("chat-example-com", "lastRooms") → "chatto:i:chat-example-com:lastRooms"
 */
export function instanceStorageKey(instanceId: string, key: string): string {
	return `chatto:i:${instanceId}:${key}`;
}

/**
 * Migrate a legacy (un-namespaced) localStorage key to a namespaced key.
 * No-op if the new key already exists or the old key is absent.
 *
 * @param instanceId - The instance to migrate to
 * @param legacyKey - The old un-namespaced key (e.g., "chatto:lastRooms")
 * @param newKeySuffix - The suffix for the new key (e.g., "lastRooms")
 */
export function migrateStorageKey(instanceId: string, legacyKey: string, newKeySuffix: string): void {
	try {
		const newKey = instanceStorageKey(instanceId, newKeySuffix);

		// Don't overwrite if namespaced key already exists
		if (localStorage.getItem(newKey) !== null) {
			return;
		}

		const oldValue = localStorage.getItem(legacyKey);
		if (oldValue === null) {
			return;
		}

		// Copy to new key, remove old key
		localStorage.setItem(newKey, oldValue);
		localStorage.removeItem(legacyKey);
	} catch {
		// Ignore storage errors (quota, security, etc.)
	}
}
