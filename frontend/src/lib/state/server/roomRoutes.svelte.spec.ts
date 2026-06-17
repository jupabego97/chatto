import { describe, expect, it, vi } from 'vitest';
import type { Client } from '@urql/svelte';
import { RoomType } from '$lib/gql/graphql';
import { NotificationLevelStore } from './notificationLevel.svelte';
import { RoomUnreadStore } from './roomUnread.svelte';
import { RoomsStore } from './rooms.svelte';
import { RoomRouteResolverStore } from './roomRoutes.svelte';

function operationName(document: unknown): string | undefined {
  return (
    document as {
      definitions?: Array<{ name?: { value?: string } }>;
    }
  ).definitions?.[0]?.name?.value;
}

function makeRoomsStore(client: Client) {
  return new RoomsStore(client, new NotificationLevelStore(), new RoomUnreadStore());
}

describe('RoomRouteResolverStore', () => {
  it('resolves loaded current room names without querying', async () => {
    const query = vi.fn();
    const client = { query } as unknown as Client;
    const rooms = makeRoomsStore(client);
    rooms.rooms = [
      {
        id: 'R1',
        name: 'General',
        type: RoomType.Channel,
        hasUnread: false,
        viewerNotificationCount: 0,
        members: []
      }
    ];

    const resolver = new RoomRouteResolverStore(client, rooms);
    await expect(resolver.resolve('general')).resolves.toMatchObject({
      roomId: 'R1',
      canonicalSegment: 'General'
    });
    expect(query).not.toHaveBeenCalled();
  });

  it('uses the focused room name query for historical aliases', async () => {
    const query = vi.fn((document: unknown) => {
      expect(operationName(document)).toBe('ResolveRoomByName');
      return {
        toPromise: () =>
          Promise.resolve({
            data: {
              roomByName: { id: 'R1', name: 'CurrentName', type: RoomType.Channel }
            },
            error: null
          })
      };
    });
    const client = { query } as unknown as Client;
    const resolver = new RoomRouteResolverStore(client, makeRoomsStore(client));

    await expect(resolver.resolve('old-name')).resolves.toMatchObject({
      roomId: 'R1',
      canonicalSegment: 'CurrentName'
    });
  });

  it('uses the legacy room query for ID-shaped segments', async () => {
    const query = vi.fn((document: unknown) => {
      expect(operationName(document)).toBe('ResolveRoomByIDFallback');
      return {
        toPromise: () =>
          Promise.resolve({
            data: { room: { id: 'R1', name: 'General', type: RoomType.Channel } },
            error: null
          })
      };
    });
    const client = { query } as unknown as Client;
    const resolver = new RoomRouteResolverStore(client, makeRoomsStore(client));

    await expect(resolver.resolve('R1')).resolves.toMatchObject({
      roomId: 'R1',
      canonicalSegment: 'General'
    });
    expect(query).toHaveBeenCalledTimes(1);
  });

  it('skips the legacy room query for non-ID-shaped unsupported names', async () => {
    const query = vi.fn((document: unknown) => {
      expect(operationName(document)).toBe('ResolveRoomByName');
      return {
        toPromise: () =>
          Promise.resolve({
            data: null,
            error: { message: 'Cannot query field "roomByName" on type "Query".' }
          })
      };
    });
    const client = { query } as unknown as Client;
    const resolver = new RoomRouteResolverStore(client, makeRoomsStore(client));

    await expect(resolver.resolve('old-name')).resolves.toBeNull();
    expect(query).toHaveBeenCalledTimes(1);
  });

  it('tries room names after an ID-shaped segment is not a room ID', async () => {
    const query = vi.fn((document: unknown) => {
      if (operationName(document) === 'ResolveRoomByIDFallback') {
        return {
          toPromise: () =>
            Promise.resolve({
              data: null,
              error: { message: 'not found' }
            })
        };
      }
      expect(operationName(document)).toBe('ResolveRoomByName');
      return {
        toPromise: () =>
          Promise.resolve({
            data: { roomByName: { id: 'R2', name: 'Random', type: RoomType.Channel } },
            error: null
          })
      };
    });
    const client = { query } as unknown as Client;
    const resolver = new RoomRouteResolverStore(client, makeRoomsStore(client));

    await expect(resolver.resolve('Random')).resolves.toMatchObject({
      roomId: 'R2',
      canonicalSegment: 'Random'
    });
  });
});
