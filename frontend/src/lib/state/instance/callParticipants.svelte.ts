/**
 * Tracks participants in an active voice call for a specific room.
 *
 * Used by the observer mode of VoiceCallPanel to show who's in a call
 * to room members who haven't joined yet.
 *
 * Data sources:
 * - Initial load: `callParticipants` GraphQL query (from CALL_STATE KV)
 * - Real-time updates: Optimistic adds/removes from CallParticipantJoined/Left events
 */

import { graphql } from '$lib/gql';
import type { CallParticipant, UserAvatarUserFragment } from '$lib/gql/graphql';
import type { Client } from '@urql/svelte';

const CallParticipantsQuery = graphql(`
	query GetCallParticipants($spaceId: ID!, $roomId: ID!) {
		callParticipants(spaceId: $spaceId, roomId: $roomId) {
			userId
			displayName
			login
			avatarUrl
			joinedAt
		}
	}
`);

/** Participant info stored in observer mode. */
export type ObserverParticipant = {
	userId: string;
	displayName: string;
	login: string;
	avatarUrl: string | null;
};

export class CallParticipantsState {
	#client: Client;

	/** Current participants visible to observers. */
	participants = $state<ObserverParticipant[]>([]);

	/** The room these participants are for. */
	private currentSpaceId: string | null = null;
	private currentRoomId: string | null = null;

	constructor(client: Client) {
		this.#client = client;
	}

	/**
	 * Load participants from the server for a specific room.
	 * Called when entering a room that has an active call.
	 */
	async load(spaceId: string, roomId: string): Promise<void> {
		this.currentSpaceId = spaceId;
		this.currentRoomId = roomId;

		const result = await this.#client
			.query(CallParticipantsQuery, { spaceId, roomId })
			.toPromise();

		if (result.data?.callParticipants) {
			this.participants = result.data.callParticipants.map(toObserverParticipant);
		}
	}

	/**
	 * Optimistically add a participant from a CallParticipantJoinedEvent.
	 * Uses the actor data from the SpaceEvent wrapper.
	 */
	handleJoin(spaceId: string, roomId: string, actor: UserAvatarUserFragment | null): void {
		if (spaceId !== this.currentSpaceId || roomId !== this.currentRoomId) return;
		if (!actor) return;

		// Avoid duplicates
		if (this.participants.some((p) => p.userId === actor.id)) return;

		this.participants = [
			...this.participants,
			{
				userId: actor.id,
				displayName: actor.displayName,
				login: actor.login,
				avatarUrl: actor.avatarUrl ?? null
			}
		];
	}

	/**
	 * Optimistically remove a participant from a CallParticipantLeftEvent.
	 */
	handleLeave(spaceId: string, roomId: string, actorId: string | null): void {
		if (spaceId !== this.currentSpaceId || roomId !== this.currentRoomId) return;
		if (!actorId) return;

		this.participants = this.participants.filter((p) => p.userId !== actorId);
	}

	/** Clear state (e.g., when leaving a room or call ends). */
	clear(): void {
		this.participants = [];
		this.currentSpaceId = null;
		this.currentRoomId = null;
	}
}

function toObserverParticipant(p: CallParticipant): ObserverParticipant {
	return {
		userId: p.userId,
		displayName: p.displayName,
		login: p.login,
		avatarUrl: p.avatarUrl ?? null
	};
}

