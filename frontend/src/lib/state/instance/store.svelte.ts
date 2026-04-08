/**
 * Bundles all instance-scoped stores into a single class per instance.
 * Created and managed by the InstanceRegistry — do not instantiate directly.
 */

import type { Client } from '@urql/svelte';
import { CurrentUserState } from '$lib/auth/currentUser.svelte';
import { InstanceState } from './state.svelte';
import type { InstancePermissions, ViewerData } from './permissions.svelte';
import { NotificationStore } from './notifications.svelte';
import { RoomUnreadStore } from './roomUnread.svelte';
import { NotificationLevelStore } from './notificationLevel.svelte';
import { VoiceCallState } from './voiceCall.svelte';
import { CallParticipantsState } from './callParticipants.svelte';
import { ActiveCallRoomsState } from './activeCallRooms.svelte';

const EMPTY_PERMISSIONS: InstancePermissions = {
	loaded: false,
	canViewAdmin: false,
	canCreateSpace: false,
	canListSpaces: false,
	canViewDMs: false,
	canWriteDMs: false,
	canAdminViewUsers: false,
	canAdminManageUsers: false,
	canAdminViewSpaces: false,
	canAdminViewRoles: false,
	canAdminManageRoles: false,
	canAdminViewSystem: false,
	canAdminViewAudit: false
};

export class InstanceStateStore {
	readonly instanceId: string;
	readonly currentUser: CurrentUserState;
	readonly instance: InstanceState;
	readonly notifications: NotificationStore;
	readonly roomUnread: RoomUnreadStore;
	readonly notificationLevels: NotificationLevelStore;
	readonly voiceCall: VoiceCallState;
	readonly callParticipants: CallParticipantsState;
	readonly activeCallRooms: ActiveCallRoomsState;

	/** Per-instance viewer permissions (loaded by InstanceSpaceSection). */
	permissions = $state<InstancePermissions>(EMPTY_PERMISSIONS);

	constructor(instanceId: string, client: Client, isOrigin: boolean) {
		this.instanceId = instanceId;

		this.currentUser = new CurrentUserState(client, isOrigin);
		this.instance = new InstanceState(client);
		this.notifications = new NotificationStore(client);
		this.roomUnread = new RoomUnreadStore();
		this.notificationLevels = new NotificationLevelStore();
		this.voiceCall = new VoiceCallState(client);
		this.callParticipants = new CallParticipantsState(client);
		this.activeCallRooms = new ActiveCallRoomsState(client, this.voiceCall);
	}

	/** Update permissions from viewer query data. */
	setPermissions(viewer: ViewerData): void {
		this.permissions = { ...viewer, loaded: true };
	}

	/** Clean up resources. */
	dispose(): void {
		this.roomUnread.clear();
		this.notificationLevels.clear();
		this.activeCallRooms.clear();
		this.callParticipants.clear();
	}
}
