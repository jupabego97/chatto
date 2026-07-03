import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tick } from 'svelte';
import { render } from 'vitest-browser-svelte';
import { RoomEventKind } from '$lib/render/eventKinds';
import type { RoomEventView } from '$lib/render/types';
import { activeServerTestState } from '$lib/test-utils/activeServerTestState.svelte';
import ThreadPane from './ThreadPane.svelte';

type ThreadFollowChanged = {
  roomId: string;
  threadRootEventId: string;
  isFollowing: boolean;
};

const mocks = vi.hoisted(() => ({
  followThread: vi.fn(),
  unfollowThread: vi.fn(),
  markThreadAsRead: vi.fn(),
  dismissThreadNotifications: vi.fn(),
  decrementUnreadNotification: vi.fn(),
  refreshNotificationCounts: vi.fn(),
  threadFollowHandlers: new Set<(update: ThreadFollowChanged) => void>()
}));

function threadRootEvent(roomId: string, threadRootEventId: string, isFollowing: boolean) {
  return {
    id: threadRootEventId,
    createdAt: '2026-06-17T10:47:00Z',
    actorId: 'root-author',
    actor: null,
    event: {
      kind: RoomEventKind.MessagePosted,
      roomId,
      body: 'thread root',
      attachments: [],
      linkPreview: null,
      reactions: [],
      updatedAt: null,
      inReplyTo: null,
      threadRootEventId: null,
      echoOfEventId: null,
      echoFromThreadRootEventId: null,
      channelEchoEventId: null,
      replyCount: 1,
      lastReplyAt: null,
      threadParticipants: [],
      viewerIsFollowingThread: isFollowing
    }
  } satisfies RoomEventView;
}

function emptyTimelinePage(roomId: string, threadRootEventId: string, isFollowing: boolean) {
  return {
    events: [threadRootEvent(roomId, threadRootEventId, isFollowing) as never],
    startCursor: null,
    endCursor: null,
    hasOlder: false,
    hasNewer: false
  };
}

function emitThreadFollow(update: ThreadFollowChanged): void {
  for (const handler of mocks.threadFollowHandlers) {
    handler(update);
  }
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

vi.mock('$lib/state/server/activeServerScope.svelte', async () => {
  const { activeServerTestState } = await import('$lib/test-utils/activeServerTestState.svelte');
  return {
    useActiveServerScope: () => ({
      get id() {
        return activeServerTestState.id;
      },
      get connection() {
        return {
          serverId: activeServerTestState.id,
          connectBaseUrl: `https://${activeServerTestState.id}.example.test/api/connect`,
          bearerToken: null
        };
      },
      get store() {
        return {
          rooms: {
            decrementUnreadNotification: mocks.decrementUnreadNotification,
            refreshNotificationCounts: mocks.refreshNotificationCounts
          }
        };
      },
      get notifications() {
        return {
          dismissThreadNotifications: mocks.dismissThreadNotifications
        };
      },
      get currentUser() {
        return { user: { id: 'viewer', login: 'viewer' }, loading: false };
      }
    })
  };
});

vi.mock('$lib/api-client/roomTimeline', () => ({
  createRoomTimelineAPI: (config: { serverId?: string }) => ({
    getRoomEvents: vi.fn(),
    getRoomEventsAround: vi.fn(),
    resolveMessageLinkTarget: vi.fn(),
    getThreadEvents: vi.fn(
      async ({ roomId, threadRootEventId }: { roomId: string; threadRootEventId: string }) =>
        emptyTimelinePage(roomId, threadRootEventId, config.serverId === 'server-b')
    ),
    getThreadEventsAround: vi.fn()
  })
}));

vi.mock('$lib/api-client/threads', () => ({
  createThreadAPI: () => ({
    followThread: mocks.followThread,
    unfollowThread: mocks.unfollowThread
  })
}));

vi.mock('$lib/api-client/readState', () => ({
  createReadStateAPI: () => ({
    markThreadAsRead: mocks.markThreadAsRead
  })
}));

vi.mock('$lib/eventBus.svelte', () => ({
  onThreadFollowChanged: (handler: (update: ThreadFollowChanged) => void) => {
    mocks.threadFollowHandlers.add(handler);
    return () => mocks.threadFollowHandlers.delete(handler);
  }
}));

vi.mock('$lib/hooks', () => ({
  useEvent: vi.fn(),
  createTypingIndicator: () => ({
    userIds: [],
    sendTypingIndicator: vi.fn(),
    resetDebounce: vi.fn(),
    removeTypingUser: vi.fn()
  })
}));

vi.mock('$lib/state/room', async (importOriginal) => {
  const actual = await importOriginal<typeof import('$lib/state/room')>();
  return {
    ...actual,
    getRoomMembers: () => []
  };
});

vi.mock('$lib/state/globals.svelte', () => ({
  appState: {
    isFocused: false,
    isPresent: true
  }
}));

vi.mock('$lib/components/composer/MessageComposer.svelte', async () => {
  const { default: EmptyMock } = await import('./RoomLocalEchoEmptyMock.svelte');
  return { default: EmptyMock };
});

vi.mock('./EventList.svelte', async () => {
  const { default: EmptyMock } = await import('./RoomLocalEchoEmptyMock.svelte');
  return { default: EmptyMock };
});

beforeEach(() => {
  activeServerTestState.id = 'server-a';
  mocks.followThread.mockReset();
  mocks.unfollowThread.mockReset();
  mocks.markThreadAsRead.mockResolvedValue({ previousReadAt: null });
  mocks.dismissThreadNotifications.mockResolvedValue({ byRoom: {} });
  mocks.decrementUnreadNotification.mockReset();
  mocks.refreshNotificationCounts.mockReset();
  mocks.threadFollowHandlers.clear();
});

describe('ThreadPane follow state', () => {
  it('ignores live follow events for the same thread id in another room', async () => {
    const rendered = render(ThreadPane, {
      props: {
        roomId: 'room-1',
        roomName: 'General',
        threadRootEventId: 'thread-1',
        onClose: vi.fn()
      }
    });

    await expect
      .element(rendered.getByRole('button', { name: 'Follow thread' }))
      .toBeInTheDocument();

    emitThreadFollow({ roomId: 'room-2', threadRootEventId: 'thread-1', isFollowing: true });
    await tick();

    await expect
      .element(rendered.getByRole('button', { name: 'Follow thread' }))
      .toBeInTheDocument();

    emitThreadFollow({ roomId: 'room-1', threadRootEventId: 'thread-1', isFollowing: true });
    await expect
      .element(rendered.getByRole('button', { name: 'Unfollow thread' }))
      .toBeInTheDocument();
  });

  it('does not let a failed follow request roll back a later server scope', async () => {
    const pendingFollow = deferred<unknown>();
    mocks.followThread.mockReturnValueOnce(pendingFollow.promise);

    const rendered = render(ThreadPane, {
      props: {
        roomId: 'room-1',
        roomName: 'General',
        threadRootEventId: 'thread-1',
        onClose: vi.fn()
      }
    });

    const followButton = rendered.getByRole('button', { name: 'Follow thread' });
    await expect.element(followButton).toBeInTheDocument();
    await followButton.click();
    await expect
      .element(rendered.getByRole('button', { name: 'Unfollow thread' }))
      .toBeInTheDocument();

    activeServerTestState.id = 'server-b';
    await rendered.rerender({
      roomId: 'room-1',
      roomName: 'General',
      threadRootEventId: 'thread-1',
      onClose: vi.fn()
    });
    await expect
      .element(rendered.getByRole('button', { name: 'Unfollow thread' }))
      .toBeInTheDocument();

    pendingFollow.reject(new Error('stale follow failed'));
    await tick();

    await expect
      .element(rendered.getByRole('button', { name: 'Unfollow thread' }))
      .toBeInTheDocument();
  });
});
