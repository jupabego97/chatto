import type { Client } from '@urql/svelte';
import { graphql } from '$lib/gql';
import { isUnsupportedGraphQLFieldError } from '$lib/gql/compatibility';
import {
  looksLikeRoomIDSegment,
  roomURLRouteKind,
  roomURLSegment,
  type RoomRouteKind,
  type RoomURLTarget
} from '$lib/roomUrls';
import type { ResolvedLoadedRoomSegment, RoomsStore } from './rooms.svelte';

const ResolveRoomByNameQuery = graphql(`
  query ResolveRoomByName($name: String!) {
    roomByName(name: $name) {
      id
      name
      type
    }
  }
`);

const ResolveRoomByIDFallbackQuery = graphql(`
  query ResolveRoomByIDFallback($roomId: ID!) {
    room(roomId: $roomId) {
      id
      name
      type
    }
  }
`);

export type ResolvedRoomRoute = ResolvedLoadedRoomSegment;

function resolvedRoom(room: RoomURLTarget): ResolvedRoomRoute {
  return {
    room: {
      id: room.id,
      name: room.name,
      type: room.type as ResolvedLoadedRoomSegment['room']['type'],
      hasUnread: false,
      viewerNotificationCount: 0,
      members: []
    },
    roomId: room.id,
    canonicalSegment: roomURLSegment(room),
    canonicalRouteKind: roomURLRouteKind(room)
  };
}

export class RoomRouteResolverStore {
  constructor(
    private readonly client: Client,
    private readonly rooms: RoomsStore
  ) {}

  async resolve(
    segment: string,
    routeKind: RoomRouteKind = 'legacy-id'
  ): Promise<ResolvedRoomRoute | null> {
    const loaded = this.rooms.resolveLoadedURLSegment(segment, routeKind);
    if (loaded) return loaded;

    if (routeKind === 'legacy-id') {
      return this.resolveLegacyID(segment);
    }

    const result = await this.client
      .query(ResolveRoomByNameQuery, { name: segment }, { requestPolicy: 'network-only' })
      .toPromise();

    if (result.error) {
      if (isUnsupportedGraphQLFieldError(result.error, 'roomByName')) return null;
      return null;
    }

    const room = result.data?.roomByName;
    return room ? resolvedRoom(room) : null;
  }

  private async resolveLegacyID(segment: string): Promise<ResolvedRoomRoute | null> {
    const loaded = this.rooms.resolveLoadedURLSegment(segment, 'legacy-id');
    if (loaded) return loaded;
    if (!looksLikeRoomIDSegment(segment)) return null;

    const result = await this.client
      .query(
        ResolveRoomByIDFallbackQuery,
        { roomId: segment },
        { requestPolicy: 'network-only' }
      )
      .toPromise();

    if (result.error) return null;
    const room = result.data?.room;
    return room ? resolvedRoom(room) : null;
  }
}
