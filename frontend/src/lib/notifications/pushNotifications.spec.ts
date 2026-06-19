import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ensureRegistered, onNotificationClick, subscribe, unsubscribe } from './pushNotifications';

const mocks = vi.hoisted(() => ({
  subscribeToPush: vi.fn(),
  unsubscribeFromPush: vi.fn()
}));

vi.mock('$lib/state/server/wireEventBus.svelte', () => ({
  wireEventBusManager: {
    getClient: () => ({
      subscribeToPush: mocks.subscribeToPush,
      unsubscribeFromPush: mocks.unsubscribeFromPush
    })
  }
}));

type TestPushSubscription = PushSubscription & {
  unsubscribe: ReturnType<typeof vi.fn>;
};

let permission: NotificationPermission;
let requestPermission: ReturnType<typeof vi.fn>;
let getSubscription: ReturnType<typeof vi.fn>;
let browserSubscribe: ReturnType<typeof vi.fn>;

function makeSubscription(endpoint: string): TestPushSubscription {
  return {
    endpoint,
    toJSON: () => ({
      endpoint,
      keys: {
        p256dh: 'p256dh-key',
        auth: 'auth-secret'
      }
    }),
    unsubscribe: vi.fn().mockResolvedValue(true)
  } as unknown as TestPushSubscription;
}

function installPushGlobals() {
  requestPermission = vi.fn(async () => {
    permission = 'granted';
    return permission;
  });
  getSubscription = vi.fn();
  browserSubscribe = vi.fn();

  vi.stubGlobal('Notification', {
    get permission() {
      return permission;
    },
    requestPermission
  });
  vi.stubGlobal('window', {
    Notification,
    PushManager: class PushManager {},
    atob: (value: string) => Buffer.from(value, 'base64').toString('binary')
  });
  vi.stubGlobal('navigator', {
    serviceWorker: {
      ready: Promise.resolve({
        pushManager: {
          getSubscription,
          subscribe: browserSubscribe
        }
      })
    },
    userAgent: 'test-agent'
  });
}

function stubServiceWorker() {
  const listeners = new Set<(event: MessageEvent) => void>();

  vi.stubGlobal('navigator', {
    serviceWorker: {
      addEventListener: vi.fn((type: string, listener: (event: MessageEvent) => void) => {
        if (type === 'message') listeners.add(listener);
      }),
      removeEventListener: vi.fn((type: string, listener: (event: MessageEvent) => void) => {
        if (type === 'message') listeners.delete(listener);
      })
    }
  });

  return {
    dispatchMessage(event: Pick<MessageEvent, 'data'>) {
      for (const listener of listeners) {
        listener(event as MessageEvent);
      }
    },
    listenerCount() {
      return listeners.size;
    }
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('pushNotifications wire registration', () => {
  beforeEach(() => {
    permission = 'default';
    installPushGlobals();
    mocks.subscribeToPush.mockReset();
    mocks.unsubscribeFromPush.mockReset();
    mocks.subscribeToPush.mockResolvedValue({ subscribed: true });
    mocks.unsubscribeFromPush.mockResolvedValue({});
  });

  it('does not prompt or subscribe when permission is default and prompt is false', async () => {
    getSubscription.mockResolvedValue(null);

    await expect(ensureRegistered('origin', 'dmFwaWQ', { prompt: false })).resolves.toBe(false);

    expect(requestPermission).not.toHaveBeenCalled();
    expect(getSubscription).not.toHaveBeenCalled();
    expect(browserSubscribe).not.toHaveBeenCalled();
    expect(mocks.subscribeToPush).not.toHaveBeenCalled();
  });

  it('saves an existing subscription when permission is granted', async () => {
    permission = 'granted';
    const existing = makeSubscription('https://push.example/existing');
    getSubscription.mockResolvedValue(existing);

    await expect(ensureRegistered('origin', 'dmFwaWQ', { prompt: false })).resolves.toBe(true);

    expect(browserSubscribe).not.toHaveBeenCalled();
    expect(mocks.subscribeToPush).toHaveBeenCalledOnce();
    expect(mocks.subscribeToPush.mock.calls[0]?.[0]).toMatchObject({
      endpoint: 'https://push.example/existing',
      p256dh: 'p256dh-key',
      auth: 'auth-secret',
      userAgent: 'test-agent'
    });
  });

  it('creates and saves a subscription when permission is granted and none exists', async () => {
    permission = 'granted';
    const created = makeSubscription('https://push.example/created');
    getSubscription.mockResolvedValue(null);
    browserSubscribe.mockResolvedValue(created);

    await expect(ensureRegistered('origin', 'dmFwaWQ', { prompt: false })).resolves.toBe(true);

    expect(browserSubscribe).toHaveBeenCalledWith({
      userVisibleOnly: true,
      applicationServerKey: expect.any(Uint8Array)
    });
    expect(mocks.subscribeToPush.mock.calls[0]?.[0]).toMatchObject({
      endpoint: 'https://push.example/created'
    });
  });

  it('prompts during explicit enable when permission is default', async () => {
    const created = makeSubscription('https://push.example/prompted');
    getSubscription.mockResolvedValue(null);
    browserSubscribe.mockResolvedValue(created);

    await expect(ensureRegistered('origin', 'dmFwaWQ', { prompt: true })).resolves.toBe(true);

    expect(requestPermission).toHaveBeenCalledOnce();
    expect(browserSubscribe).toHaveBeenCalledOnce();
    expect(mocks.subscribeToPush).toHaveBeenCalledOnce();
  });

  it('cleans up only newly created subscriptions when server save fails', async () => {
    permission = 'granted';
    const existing = makeSubscription('https://push.example/existing');
    getSubscription.mockResolvedValueOnce(existing);
    mocks.subscribeToPush.mockResolvedValueOnce({ subscribed: false });

    await expect(ensureRegistered('origin', 'dmFwaWQ', { prompt: false })).resolves.toBe(false);
    expect(existing.unsubscribe).not.toHaveBeenCalled();

    const created = makeSubscription('https://push.example/created');
    getSubscription.mockResolvedValueOnce(null);
    browserSubscribe.mockResolvedValueOnce(created);
    mocks.subscribeToPush.mockResolvedValueOnce({ subscribed: false });

    await expect(ensureRegistered('origin', 'dmFwaWQ', { prompt: false })).resolves.toBe(false);
    expect(created.unsubscribe).toHaveBeenCalledOnce();
  });

  it('subscribes explicitly through the wire API', async () => {
    const created = makeSubscription('https://push.example/subscribe');
    browserSubscribe.mockResolvedValue(created);

    await expect(subscribe('origin', 'dmFwaWQ')).resolves.toBe(true);

    expect(requestPermission).toHaveBeenCalledOnce();
    expect(mocks.subscribeToPush.mock.calls[0]?.[0]).toMatchObject({
      endpoint: 'https://push.example/subscribe'
    });
  });

  it('unsubscribes through the wire API before removing the browser subscription', async () => {
    const existing = makeSubscription('https://push.example/remove');
    getSubscription.mockResolvedValue(existing);

    await expect(unsubscribe('origin')).resolves.toBe(true);

    expect(mocks.unsubscribeFromPush.mock.calls[0]?.[0]).toMatchObject({
      endpoint: 'https://push.example/remove'
    });
    expect(existing.unsubscribe).toHaveBeenCalledOnce();
  });
});

describe('onNotificationClick', () => {
  it('calls the notification navigation callback', () => {
    const serviceWorker = stubServiceWorker();
    const callback = vi.fn();
    const stop = onNotificationClick(callback);

    serviceWorker.dispatchMessage({
      data: {
        type: 'notification-click',
        url: 'https://chatto.example/chat/-/room-1'
      }
    });

    expect(callback).toHaveBeenCalledWith('https://chatto.example/chat/-/room-1');

    stop();
    expect(serviceWorker.listenerCount()).toBe(0);
  });
});
