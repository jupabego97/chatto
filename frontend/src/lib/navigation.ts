import { instanceRegistry } from '$lib/state/instance/registry.svelte';

/** URL segment used for the home (origin) instance. */
const HOME_SEGMENT = '-';

/**
 * Convert an internal instance registry ID to a URL segment.
 * Origin instance → "-", remote → raw hostname from URL.
 */
export function instanceIdToSegment(instanceId: string): string {
	if (instanceRegistry.isOriginInstance(instanceId)) return HOME_SEGMENT;

	const instance = instanceRegistry.getInstance(instanceId);
	if (!instance) return HOME_SEGMENT;

	try {
		return new URL(instance.url).hostname;
	} catch {
		return HOME_SEGMENT;
	}
}

/**
 * Convert a URL segment back to an internal instance registry ID.
 * "-" → origin instance, hostname → find matching instance by URL.
 */
export function segmentToInstanceId(segment: string): string | null {
	if (segment === HOME_SEGMENT) {
		return instanceRegistry.originInstance?.id ?? null;
	}

	// Find instance whose URL hostname matches the segment
	for (const instance of instanceRegistry.instances) {
		try {
			if (new URL(instance.url).hostname === segment) {
				return instance.id;
			}
		} catch {
			continue;
		}
	}

	return null;
}
