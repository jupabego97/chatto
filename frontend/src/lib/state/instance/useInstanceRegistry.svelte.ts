import { instanceRegistry } from './registry.svelte';

/**
 * Bootstrap the instance registry: create stores, probe the origin,
 * and re-fetch instance info when the tab resumes.
 *
 * Must be called during component initialization (root layout script).
 *
 * @param getUser - Getter returning the current user (truthy = known instance,
 *   falsy = probe needed). Passed as a getter so reads happen inside `$effect`.
 */
export function useInstanceRegistry(getUser: () => unknown): void {
	instanceRegistry.init();
	instanceRegistry.probeOrigin(!!getUser());

	// Re-fetch instance info when the tab becomes visible
	$effect(() => {
		const originId = instanceRegistry.originInstance?.id;
		if (!originId) return;

		const onVisibilityChange = () => {
			if (document.visibilityState === 'visible') {
				instanceRegistry.getStore(originId).instance.init();
			}
		};

		document.addEventListener('visibilitychange', onVisibilityChange);
		return () => document.removeEventListener('visibilitychange', onVisibilityChange);
	});
}
