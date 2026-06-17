import { resolve } from '$app/paths';
import type { ResolvedPathname } from '$app/types';
import { RoomType } from '$lib/gql/graphql';
import { serverIdToSegment } from '$lib/navigation';

export type RoomURLTarget = {
  id: string;
  name: string;
  type?: RoomType | string | null;
};

export type RoomRouteKind = 'legacy-id' | 'name';

export function looksLikeRoomIDSegment(segment: string): boolean {
  return (
    (segment.startsWith('R') && /^[A-Za-z0-9_-]{1,30}$/.test(segment)) ||
    /^[a-f0-9]{14}$/.test(segment)
  );
}

export function roomURLSegment(room: RoomURLTarget): string {
  if (room.type === RoomType.Dm || room.type === 'DM') {
    return room.id;
  }
  const name = room.name.trim();
  return name || room.id;
}

export function roomURLRouteKind(room: RoomURLTarget): RoomRouteKind {
  if (room.type === RoomType.Dm || room.type === 'DM') {
    return 'legacy-id';
  }
  return room.name.trim() ? 'name' : 'legacy-id';
}

export function roomPathForTarget(
  serverSegment: string,
  room: RoomURLTarget
): ResolvedPathname {
  return roomPathForSegment(serverSegment, roomURLSegment(room), roomURLRouteKind(room));
}

export function roomPath(serverId: string, room: RoomURLTarget): ResolvedPathname {
  return roomPathForTarget(serverIdToSegment(serverId), room);
}

export function roomThreadPath(
  serverId: string,
  room: RoomURLTarget,
  threadId: string
): ResolvedPathname {
  return roomThreadPathForSegment(
    serverIdToSegment(serverId),
    roomURLSegment(room),
    threadId,
    roomURLRouteKind(room)
  );
}

export function roomMessagePath(
  serverId: string,
  room: RoomURLTarget,
  messageId: string
): ResolvedPathname {
  return roomMessagePathForSegment(
    serverIdToSegment(serverId),
    roomURLSegment(room),
    messageId,
    roomURLRouteKind(room)
  );
}

export function roomPathForSegment(
  serverSegment: string,
  roomSegment: string,
  routeKind: RoomRouteKind = 'legacy-id'
): ResolvedPathname {
  if (routeKind === 'name') {
    return resolve('/chat/[serverId]/r/[roomId]', {
      serverId: serverSegment,
      roomId: roomSegment
    });
  }
  return resolve('/chat/[serverId]/[roomId]', {
    serverId: serverSegment,
    roomId: roomSegment
  });
}

export function roomThreadPathForSegment(
  serverSegment: string,
  roomSegment: string,
  threadId: string,
  routeKind: RoomRouteKind = 'legacy-id'
): ResolvedPathname {
  if (routeKind === 'name') {
    return resolve('/chat/[serverId]/r/[roomId]/[threadId]', {
      serverId: serverSegment,
      roomId: roomSegment,
      threadId
    });
  }
  return resolve('/chat/[serverId]/[roomId]/[threadId]', {
    serverId: serverSegment,
    roomId: roomSegment,
    threadId
  });
}

export function roomMessagePathForSegment(
  serverSegment: string,
  roomSegment: string,
  messageId: string,
  routeKind: RoomRouteKind = 'legacy-id'
): ResolvedPathname {
  if (routeKind === 'name') {
    return resolve('/chat/[serverId]/r/[roomId]/m/[messageId]', {
      serverId: serverSegment,
      roomId: roomSegment,
      messageId
    });
  }
  return resolve('/chat/[serverId]/[roomId]/m/[messageId]', {
    serverId: serverSegment,
    roomId: roomSegment,
    messageId
  });
}
