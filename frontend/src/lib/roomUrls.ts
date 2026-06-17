import { resolve } from '$app/paths';
import type { ResolvedPathname } from '$app/types';
import { RoomType } from '$lib/gql/graphql';
import { serverIdToSegment } from '$lib/navigation';

export type RoomURLTarget = {
  id: string;
  name: string;
  type?: RoomType | string | null;
};

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

export function roomPath(serverId: string, room: RoomURLTarget): ResolvedPathname {
  return roomPathForSegment(serverIdToSegment(serverId), roomURLSegment(room));
}

export function roomThreadPath(
  serverId: string,
  room: RoomURLTarget,
  threadId: string
): ResolvedPathname {
  return roomThreadPathForSegment(serverIdToSegment(serverId), roomURLSegment(room), threadId);
}

export function roomMessagePath(
  serverId: string,
  room: RoomURLTarget,
  messageId: string
): ResolvedPathname {
  return roomMessagePathForSegment(serverIdToSegment(serverId), roomURLSegment(room), messageId);
}

export function roomPathForSegment(
  serverSegment: string,
  roomSegment: string
): ResolvedPathname {
  return resolve('/chat/[serverId]/[roomId]', {
    serverId: serverSegment,
    roomId: roomSegment
  });
}

export function roomThreadPathForSegment(
  serverSegment: string,
  roomSegment: string,
  threadId: string
): ResolvedPathname {
  return resolve('/chat/[serverId]/[roomId]/[threadId]', {
    serverId: serverSegment,
    roomId: roomSegment,
    threadId
  });
}

export function roomMessagePathForSegment(
  serverSegment: string,
  roomSegment: string,
  messageId: string
): ResolvedPathname {
  return resolve('/chat/[serverId]/[roomId]/m/[messageId]', {
    serverId: serverSegment,
    roomId: roomSegment,
    messageId
  });
}
