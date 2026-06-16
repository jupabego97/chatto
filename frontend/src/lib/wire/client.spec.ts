import { describe, expect, it } from 'vitest';
import {
  ClientFrame,
  ErrorCode,
  ServerFrame,
  ServerHello,
  StreamEvent,
  WireError,
  Response as WireResponse
} from '$lib/pb/chatto/wire/v1/protocol_pb';
import { GetViewerRequest, GetViewerResponse, Viewer } from '$lib/pb/chatto/api/v1/chat_pb';
import { HeartbeatEvent } from '$lib/pb/chatto/core/v1/live_events_pb';
import { User } from '$lib/pb/chatto/core/v1/models_pb';
import { WireClient, type WireSocket, httpToWireWsUrl, wireMethods } from './client';

type FakeListener =
  | ((event: Event) => void)
  | ((event: MessageEvent) => void)
  | ((event: CloseEvent) => void);

class FakeWireSocket implements WireSocket {
  binaryType: BinaryType = 'blob';
  readyState = 0;
  readonly sent: Uint8Array[] = [];
  readonly url: string;
  #listeners = new Map<string, Set<FakeListener>>();

  constructor(url: string) {
    this.url = url;
  }

  send(data: Uint8Array): void {
    this.sent.push(data);
  }

  close(): void {
    this.readyState = 3;
    this.#emit('close', {} as CloseEvent);
  }

  addEventListener(type: 'open', listener: (event: Event) => void): void;
  addEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
  addEventListener(type: 'close', listener: (event: CloseEvent) => void): void;
  addEventListener(type: 'error', listener: (event: Event) => void): void;
  addEventListener(type: string, listener: FakeListener): void {
    const listeners = this.#listeners.get(type) ?? new Set<FakeListener>();
    listeners.add(listener);
    this.#listeners.set(type, listeners);
  }

  removeEventListener(type: 'open', listener: (event: Event) => void): void;
  removeEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
  removeEventListener(type: 'close', listener: (event: CloseEvent) => void): void;
  removeEventListener(type: 'error', listener: (event: Event) => void): void;
  removeEventListener(type: string, listener: FakeListener): void {
    this.#listeners.get(type)?.delete(listener);
  }

  open(): void {
    this.readyState = 1;
    this.#emit('open', {} as Event);
  }

  serverSend(frame: ServerFrame): void {
    this.#emit('message', { data: frame.toBinary() } as MessageEvent);
  }

  lastClientFrame(): ClientFrame {
    const data = this.sent.at(-1);
    if (!data) throw new Error('fake socket did not record a client frame');
    return ClientFrame.fromBinary(data);
  }

  #emit(type: string, event: Event | MessageEvent | CloseEvent): void {
    for (const listener of this.#listeners.get(type) ?? []) {
      (listener as (event: Event | MessageEvent | CloseEvent) => void)(event);
    }
  }
}

function makeClient(config: { token?: string | null } = {}): {
  client: WireClient;
  socket(): FakeWireSocket;
} {
  let socket: FakeWireSocket | null = null;
  const client = new WireClient({
    url: '/api/wire',
    token: config.token ?? null,
    socketFactory: (url) => {
      socket = new FakeWireSocket(url);
      return socket;
    }
  });
  return {
    client,
    socket() {
      if (!socket) throw new Error('wire client has not opened a socket');
      return socket;
    }
  };
}

async function connectClient(harness: ReturnType<typeof makeClient>): Promise<ServerHello> {
  const helloPromise = harness.client.connect();
  const socket = harness.socket();
  socket.open();
  socket.serverSend(
    new ServerFrame({
      frameId: 'hello',
      kind: {
        case: 'hello',
        value: new ServerHello({
          protocolVersion: 'chatto-wire-v1',
          serverVersion: 'test',
          methods: Object.values(wireMethods),
          features: ['binary-protobuf', 'requests', 'my-events']
        })
      }
    })
  );
  return helloPromise;
}

function waitForMessageHandling(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe('httpToWireWsUrl', () => {
  it('converts HTTP(S) endpoints to WebSocket endpoints', () => {
    expect(httpToWireWsUrl('http://localhost:4000/api/wire')).toBe('ws://localhost:4000/api/wire');
    expect(httpToWireWsUrl('https://chat.example.com/api/wire')).toBe(
      'wss://chat.example.com/api/wire'
    );
    expect(httpToWireWsUrl('wss://chat.example.com/api/wire')).toBe(
      'wss://chat.example.com/api/wire'
    );
  });
});

describe('WireClient', () => {
  it('opens with a protobuf ClientHello carrying resume and bearer auth', async () => {
    const harness = makeClient({ token: 'opaque-token' });
    const { client } = harness;
    const helloPromise = client.connect({
      resumeAfter: 'cursor-123',
      acceptedFeatures: ['example-feature']
    });

    const socket = harness.socket();
    socket.open();
    const sent = socket.lastClientFrame();
    if (sent.kind.case !== 'hello') throw new Error('expected ClientHello frame');

    expect(sent.kind.value.protocolVersion).toBe('chatto-wire-v1');
    expect(sent.kind.value.resumeAfter).toBe('cursor-123');
    expect(sent.kind.value.acceptedFeatures).toEqual(['example-feature']);
    expect(sent.kind.value.bearerToken).toBe('opaque-token');

    socket.serverSend(
      new ServerFrame({
        frameId: sent.frameId,
        kind: {
          case: 'hello',
          value: new ServerHello({
            protocolVersion: 'chatto-wire-v1',
            serverVersion: 'test',
            methods: Object.values(wireMethods),
            features: ['binary-protobuf']
          })
        }
      })
    );

    const hello = await helloPromise;
    expect(hello.methods).toContain(wireMethods.getViewer);
    expect(client.status).toBe('connected');
  });

  it('sends a typed request, handles an interleaved event, and decodes the typed response', async () => {
    const harness = makeClient();
    const { client } = harness;
    await connectClient(harness);
    const socket = harness.socket();

    const events: StreamEvent[] = [];
    client.onEvent((event) => events.push(event));

    const responsePromise = client.getViewer();
    await waitForMessageHandling();
    const requestFrame = socket.lastClientFrame();
    if (requestFrame.kind.case !== 'request') throw new Error('expected Request frame');

    const request = requestFrame.kind.value;
    expect(request.method).toBe(wireMethods.getViewer);
    expect(GetViewerRequest.fromBinary(request.body)).toBeInstanceOf(GetViewerRequest);

    const streamEvent = new StreamEvent({
      eventId: 'evt_1',
      deliveryCursor: 'cursor_1',
      eventType: 'heartbeat',
      payload: {
        case: 'heartbeat',
        value: new HeartbeatEvent()
      }
    });
    socket.serverSend(
      new ServerFrame({
        kind: {
          case: 'event',
          value: streamEvent
        }
      })
    );
    await waitForMessageHandling();

    expect(events).toHaveLength(1);
    expect(events[0].eventId).toBe('evt_1');
    expect(client.lastDeliveryCursor).toBe('cursor_1');

    client.ack(events[0]);
    const ackFrame = socket.lastClientFrame();
    if (ackFrame.kind.case !== 'ack') throw new Error('expected Ack frame');
    expect(ackFrame.kind.value.deliveryCursor).toBe('cursor_1');

    socket.serverSend(
      new ServerFrame({
        frameId: requestFrame.frameId,
        kind: {
          case: 'response',
          value: new WireResponse({
            requestId: request.requestId,
            body: new Uint8Array(
              new GetViewerResponse({
                viewer: new Viewer({
                  user: new User({
                    id: 'user_1',
                    login: 'test-user',
                    displayName: 'Test User'
                  })
                })
              }).toBinary()
            )
          })
        }
      })
    );

    const response = await responsePromise;
    expect(response.viewer?.user?.id).toBe('user_1');
    expect(response.viewer?.user?.displayName).toBe('Test User');
  });

  it('rejects the matching request when the server returns a protobuf WireError', async () => {
    const harness = makeClient();
    const { client } = harness;
    await connectClient(harness);
    const socket = harness.socket();

    const responsePromise = client.request(
      '/chatto.api.v1.ChattoApiService/Nope',
      new GetViewerRequest(),
      GetViewerResponse,
      { requestId: 'req_error' }
    );
    await waitForMessageHandling();
    const requestFrame = socket.lastClientFrame();
    if (requestFrame.kind.case !== 'request') throw new Error('expected Request frame');

    socket.serverSend(
      new ServerFrame({
        frameId: requestFrame.frameId,
        kind: {
          case: 'error',
          value: new WireError({
            requestId: 'req_error',
            code: ErrorCode.UNIMPLEMENTED,
            message: 'unknown method'
          })
        }
      })
    );

    await expect(responsePromise).rejects.toMatchObject({
      name: 'WireProtocolError',
      wireError: expect.objectContaining({ code: ErrorCode.UNIMPLEMENTED })
    });
  });
});
