<script lang="ts">
	import type { Snippet } from 'svelte';
	import { provideConnection } from '$lib/state/instance/connection.svelte';
	import { graphqlClientManager } from '$lib/state/instance/graphqlClient.svelte';

	let { children }: { children: Snippet } = $props();

	// Provide the home instance's GraphQL client by default.
	// The [[instanceId=hostname]] layout overrides this with the
	// correct instance-specific client for its subtree.
	provideConnection(() => graphqlClientManager.originClient);
</script>

{@render children()}
