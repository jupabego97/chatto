import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { flushSync } from 'svelte';
import { render } from 'vitest-browser-svelte';
import { q } from '$lib/test-utils';
import { RoomEventKind } from '$lib/render/eventKinds';
import {
  eventBusManager,
  setRealtimeSocketFactoryForTests
} from '$lib/state/server/eventBus.svelte';
import type { EventEnvelope } from '$lib/eventBus.svelte';
import type { ServerConnection } from '$lib/state/server/serverConnection.svelte';
import Harness from './TypingIndicatorHarness.svelte';

const TEST_SERVER = 'test-server';

vi.mock('$lib/state/server/activeServerScope.svelte', () => ({
  useActiveServerScope: () => ({
    id: TEST_SERVER,
    connection: {
      serverId: TEST_SERVER,
      connectBaseUrl: 'https://test-server.example.test/api/connect',
      bearerToken: 'test-token'
    }
  })
}));

class FakeServerConnection {
  realtimeUrl = 'ws://test-server/api/realtime';
  bearerToken: string | null = null;
  setRealtimeConnectionStatus = vi.fn();
  registerRealtimeReconnect = vi.fn(() => () => {});
  handleAuthenticationRequired = vi.fn();
}

function typingEvent(userId: string): EventEnvelope {
  return {
    id: `typing-${userId}`,
    actorId: userId,
    createdAt: new Date().toISOString(),
    event: {
      kind: RoomEventKind.UserTyping,
      roomId: 'room-1',
      typingThreadRootEventId: null
    }
  };
}

function emitTyping(userId: string): void {
  const bus = eventBusManager.getBus(TEST_SERVER);
  if (!bus) throw new Error('event bus did not start');
  for (const handler of bus.handlers) {
    handler(typingEvent(userId));
  }
}

describe('createTypingIndicator', () => {
  let consoleDebug: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    setRealtimeSocketFactoryForTests(() => ({
      binaryType: 'arraybuffer',
      readyState: 0,
      onopen: null,
      onmessage: null,
      onerror: null,
      onclose: null,
      send: vi.fn(),
      close: vi.fn()
    }));
    consoleDebug = vi.spyOn(console, 'debug').mockImplementation(() => {});
  });

  afterEach(() => {
    eventBusManager.stopBus(TEST_SERVER);
    setRealtimeSocketFactoryForTests(null);
    consoleDebug.mockRestore();
    vi.restoreAllMocks();
  });

  it('re-subscribes when the active server event bus is replaced', async () => {
    eventBusManager.startBus(TEST_SERVER, new FakeServerConnection() as unknown as ServerConnection);

    const rendered = render(Harness, {
      props: {
        serverId: TEST_SERVER,
        roomId: 'room-1',
        currentUserId: 'current-user'
      }
    });
    flushSync();

    await vi.waitFor(() => expect(eventBusManager.getBus(TEST_SERVER)?.handlers.size).toBe(1));

    emitTyping('before-restart');
    await vi.waitFor(() =>
      expect(q(rendered.container, '[data-testid="typing-users"]')).toHaveTextContent(
        'before-restart'
      )
    );

    eventBusManager.stopBus(TEST_SERVER);
    flushSync();
    eventBusManager.startBus(TEST_SERVER, new FakeServerConnection() as unknown as ServerConnection);
    flushSync();

    await vi.waitFor(() => expect(eventBusManager.getBus(TEST_SERVER)?.handlers.size).toBe(1));

    emitTyping('after-restart');
    await vi.waitFor(() =>
      expect(q(rendered.container, '[data-testid="typing-users"]')).toHaveTextContent(
        'after-restart'
      )
    );

    rendered.unmount();
  });
});
