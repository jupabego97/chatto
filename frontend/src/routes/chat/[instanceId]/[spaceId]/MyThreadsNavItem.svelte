<script lang="ts">
	import { resolve } from '$app/paths';
	import { instanceIdToSegment } from '$lib/navigation';
	import { getActiveInstance } from '$lib/state/activeInstance.svelte';
	import { instanceRegistry } from '$lib/state/instance/registry.svelte';

	const getInstanceId = getActiveInstance();

	let { spaceId, active }: { spaceId: string; active: boolean } = $props();

	const notificationStore = instanceRegistry.getStore(getInstanceId()).notifications;

	const hasUnread = $derived(
		notificationStore.notifications.some(
			(n) =>
				n.__typename === 'ReplyNotificationItem' &&
				n.replyInThread &&
				n.replySpace?.id === spaceId
		)
	);
</script>

<a
	href={resolve('/chat/[instanceId]/[spaceId]/threads', { instanceId: instanceIdToSegment(getInstanceId()), spaceId })}
	class={['sidebar-item', active ? 'bg-surface-100' : 'text-muted']}
>
	<span class="sidebar-icon iconify uil--comment-alt-lines"></span>
	My Threads
	{#if hasUnread}
		<span
			class="ml-auto h-2 w-2 shrink-0 rounded-full bg-warning"
			data-testid="my-threads-unread-dot"
		></span>
	{/if}
</a>
