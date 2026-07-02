import { serverIdToSegment } from '$lib/navigation';
import { getActiveServer } from '$lib/state/activeServer.svelte';
import { serverRegistry } from './registry.svelte';
import { useTrackedConnection } from './connection.svelte';
import type { ServerConnection } from './serverConnection.svelte';
import type { ServerStateStore } from './store.svelte';

/**
 * Reactive facade for the URL-selected server and its per-server state.
 *
 * Route-scoped code should use this instead of open-coding an active-server
 * registry lookup. The facade also makes the contextual connection track the
 * URL server even though the low-level `useConnection()` helper intentionally
 * returns an untracked getter.
 */
export function useActiveServerScope() {
	const id = $derived(getActiveServer());
	const getConnection = useTrackedConnection(() => id);
	const connection = $derived(getConnection());
	const segment = $derived(serverIdToSegment(id));
	const store = $derived(serverRegistry.getStore(id));

	return {
		get id(): string {
			return id;
		},
		get segment(): string {
			return segment;
		},
		get connection(): ServerConnection {
			return connection;
		},
		get store(): ServerStateStore {
			return store;
		},
		get currentUser() {
			return store.currentUser;
		},
		get serverInfo() {
			return store.serverInfo;
		},
		get notifications() {
			return store.notifications;
		},
		get roomUnread() {
			return store.roomUnread;
		},
		get rooms() {
			return store.rooms;
		},
		get pendingHighlights() {
			return store.pendingHighlights;
		},
		get voiceCall() {
			return store.voiceCall;
		},
		get activeCallRooms() {
			return store.activeCallRooms;
		},
		get callParticipants() {
			return store.callParticipants;
		}
	};
}

export type ActiveServerScope = ReturnType<typeof useActiveServerScope>;
