import {
  Ack,
  ClientFrame,
  ClientHello,
  CancelRequest,
  Request,
  Response,
  ServerFrame,
  ServerHello,
  StreamEvent,
  WireError
} from '$lib/pb/chatto/wire/v1/protocol_pb';
import {
  GetRoomTimelineRequest,
  GetRoomTimelineResponse,
  GetViewerRequest,
  GetViewerResponse,
  ListMyRoomsRequest,
  ListMyRoomsResponse,
  PostMessageRequest,
  PostMessageResponse,
  SendTypingIndicatorRequest,
  SendTypingIndicatorResponse
} from '$lib/pb/chatto/api/v1/chat_pb';

const WEBSOCKET_OPEN = 1;
const DEFAULT_WIRE_URL = '/api/wire';
const WIRE_PROTOCOL_VERSION = 'chatto-wire-v1';

export const wireMethods = {
  getViewer: '/chatto.api.v1.ChattoApiService/GetViewer',
  listMyRooms: '/chatto.api.v1.ChattoApiService/ListMyRooms',
  getRoomTimeline: '/chatto.api.v1.ChattoApiService/GetRoomTimeline',
  postMessage: '/chatto.api.v1.ChattoApiService/PostMessage',
  sendTypingIndicator: '/chatto.api.v1.ChattoApiService/SendTypingIndicator'
} as const;

export type WireMethod = (typeof wireMethods)[keyof typeof wireMethods];

export type WireConnectionStatus = 'idle' | 'connecting' | 'connected' | 'closed';

export interface WireSocket {
  binaryType: BinaryType;
  readyState: number;
  send(data: Uint8Array): void;
  close(code?: number, reason?: string): void;
  addEventListener(type: 'open', listener: (event: Event) => void): void;
  addEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
  addEventListener(type: 'close', listener: (event: CloseEvent) => void): void;
  addEventListener(type: 'error', listener: (event: Event) => void): void;
  removeEventListener(type: 'open', listener: (event: Event) => void): void;
  removeEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
  removeEventListener(type: 'close', listener: (event: CloseEvent) => void): void;
  removeEventListener(type: 'error', listener: (event: Event) => void): void;
}

export type WireSocketFactory = (url: string) => WireSocket;

export interface WireClientConfig {
  url?: string;
  token?: string | null;
  socketFactory?: WireSocketFactory;
}

export interface WireConnectOptions {
  resumeAfter?: string;
  acceptedFeatures?: string[];
}

export interface WireRequestOptions {
  requestId?: string;
}

interface BinaryMessage {
  toBinary(): Uint8Array<ArrayBufferLike>;
}

interface BinaryMessageType<T> {
  fromBinary(bytes: Uint8Array): T;
}

interface PendingRequest<T> {
  responseType: BinaryMessageType<T>;
  resolve(value: T): void;
  reject(error: unknown): void;
}

type StreamEventListener = (event: StreamEvent) => void;
type WireErrorListener = (error: WireProtocolError) => void;

export class WireProtocolError extends Error {
  readonly wireError?: WireError;

  constructor(message: string, wireError?: WireError) {
    super(message);
    this.name = 'WireProtocolError';
    this.wireError = wireError;
  }
}

export function httpToWireWsUrl(url: string): string {
  if (url.startsWith('ws://') || url.startsWith('wss://')) return url;
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url.replace(/^http/, 'ws');
  }
  if (url.startsWith('/') && typeof window !== 'undefined') {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}${url}`;
  }
  return url;
}

export class WireClient {
  status: WireConnectionStatus = 'idle';
  lastDeliveryCursor = '';

  #url: string;
  #token: string | null;
  #socketFactory: WireSocketFactory;
  #socket: WireSocket | null = null;
  #connectPromise: Promise<ServerHello> | null = null;
  #helloResolve: ((hello: ServerHello) => void) | null = null;
  #helloReject: ((error: unknown) => void) | null = null;
  #requestSeq = 0;
  #frameSeq = 0;
  #pendingRequests = new Map<string, PendingRequest<unknown>>();
  #eventListeners = new Set<StreamEventListener>();
  #errorListeners = new Set<WireErrorListener>();

  constructor(config: WireClientConfig = {}) {
    this.#url = httpToWireWsUrl(config.url ?? DEFAULT_WIRE_URL);
    this.#token = config.token ?? null;
    this.#socketFactory = config.socketFactory ?? ((url) => new WebSocket(url));
  }

  connect(options: WireConnectOptions = {}): Promise<ServerHello> {
    if (this.#connectPromise) return this.#connectPromise;

    this.status = 'connecting';
    const socket = this.#socketFactory(this.#url);
    socket.binaryType = 'arraybuffer';
    this.#socket = socket;

    socket.addEventListener('open', this.#handleOpen(options));
    socket.addEventListener('message', this.#handleMessage);
    socket.addEventListener('close', this.#handleClose);
    socket.addEventListener('error', this.#handleSocketError);

    this.#connectPromise = new Promise<ServerHello>((resolve, reject) => {
      this.#helloResolve = resolve;
      this.#helloReject = reject;
    });
    return this.#connectPromise;
  }

  dispose(): void {
    const error = new WireProtocolError('wire connection disposed');
    this.#helloReject?.(error);
    this.#helloResolve = null;
    this.#helloReject = null;
    this.#rejectAll(error);
    this.#socket?.close();
    this.#detachSocket();
    this.status = 'closed';
  }

  onEvent(listener: StreamEventListener): () => void {
    this.#eventListeners.add(listener);
    return () => this.#eventListeners.delete(listener);
  }

  onError(listener: WireErrorListener): () => void {
    this.#errorListeners.add(listener);
    return () => this.#errorListeners.delete(listener);
  }

  async request<TResponse>(
    method: WireMethod | string,
    body: BinaryMessage,
    responseType: BinaryMessageType<TResponse>,
    options: WireRequestOptions = {}
  ): Promise<TResponse> {
    await this.connect();
    const requestId = options.requestId ?? this.#nextRequestId();
    if (this.#pendingRequests.has(requestId)) {
      throw new WireProtocolError(`wire request ${requestId} is already in flight`);
    }

    const promise = new Promise<TResponse>((resolve, reject) => {
      this.#pendingRequests.set(requestId, {
        responseType,
        resolve: resolve as (value: unknown) => void,
        reject
      });
    });

    const frame = new ClientFrame({
      frameId: this.#nextFrameId(),
      kind: {
        case: 'request',
        value: new Request({
          requestId,
          method,
          body: protobufBytes(body)
        })
      }
    });

    try {
      this.#send(frame);
    } catch (error: unknown) {
      this.#pendingRequests.delete(requestId);
      throw error;
    }

    return promise;
  }

  cancel(requestId: string): void {
    this.#send(
      new ClientFrame({
        frameId: this.#nextFrameId(),
        kind: {
          case: 'cancel',
          value: new CancelRequest({ requestId })
        }
      })
    );
  }

  ack(event: StreamEvent): void {
    this.#send(
      new ClientFrame({
        frameId: this.#nextFrameId(),
        kind: {
          case: 'ack',
          value: new Ack({
            eventId: event.eventId,
            deliveryCursor: event.deliveryCursor
          })
        }
      })
    );
  }

  getViewer(): Promise<GetViewerResponse> {
    return this.request(wireMethods.getViewer, new GetViewerRequest(), GetViewerResponse);
  }

  listMyRooms(request = new ListMyRoomsRequest()): Promise<ListMyRoomsResponse> {
    return this.request(wireMethods.listMyRooms, request, ListMyRoomsResponse);
  }

  getRoomTimeline(request: GetRoomTimelineRequest): Promise<GetRoomTimelineResponse> {
    return this.request(wireMethods.getRoomTimeline, request, GetRoomTimelineResponse);
  }

  postMessage(request: PostMessageRequest): Promise<PostMessageResponse> {
    return this.request(wireMethods.postMessage, request, PostMessageResponse);
  }

  sendTypingIndicator(request: SendTypingIndicatorRequest): Promise<SendTypingIndicatorResponse> {
    return this.request(wireMethods.sendTypingIndicator, request, SendTypingIndicatorResponse);
  }

  #handleOpen = (options: WireConnectOptions) => (): void => {
    this.#send(
      new ClientFrame({
        frameId: this.#nextFrameId(),
        kind: {
          case: 'hello',
          value: new ClientHello({
            protocolVersion: WIRE_PROTOCOL_VERSION,
            resumeAfter: options.resumeAfter ?? '',
            acceptedFeatures: options.acceptedFeatures ?? [],
            bearerToken: this.#token ?? ''
          })
        }
      })
    );
  };

  #handleMessage = (event: MessageEvent): void => {
    void this.#decodeFrame(event.data)
      .then((frame) => this.#handleServerFrame(frame))
      .catch((error: unknown) => {
        this.#emitError(new WireProtocolError(errorMessage(error)));
      });
  };

  #handleServerFrame(frame: ServerFrame): void {
    switch (frame.kind.case) {
      case 'hello':
        this.status = 'connected';
        this.#helloResolve?.(frame.kind.value);
        this.#helloResolve = null;
        this.#helloReject = null;
        return;
      case 'response':
        this.#handleResponse(frame.kind.value);
        return;
      case 'event':
        this.#handleStreamEvent(frame.kind.value);
        return;
      case 'error':
        this.#handleWireError(frame.kind.value);
        return;
      default:
        this.#emitError(new WireProtocolError('wire server frame kind is missing'));
    }
  }

  #handleResponse(response: Response): void {
    const pending = this.#pendingRequests.get(response.requestId);
    if (!pending) return;

    this.#pendingRequests.delete(response.requestId);
    try {
      pending.resolve(pending.responseType.fromBinary(response.body));
    } catch (error: unknown) {
      pending.reject(new WireProtocolError(errorMessage(error)));
    }
  }

  #handleStreamEvent(event: StreamEvent): void {
    if (event.deliveryCursor) {
      this.lastDeliveryCursor = event.deliveryCursor;
    }
    for (const listener of this.#eventListeners) listener(event);
  }

  #handleWireError(error: WireError): void {
    const protocolError = new WireProtocolError(error.message, error);
    if (error.requestId) {
      const pending = this.#pendingRequests.get(error.requestId);
      if (pending) {
        this.#pendingRequests.delete(error.requestId);
        pending.reject(protocolError);
        return;
      }
    }
    this.#helloReject?.(protocolError);
    this.#emitError(protocolError);
  }

  #handleClose = (): void => {
    this.status = 'closed';
    this.#helloReject?.(new WireProtocolError('wire connection closed before hello'));
    this.#helloResolve = null;
    this.#helloReject = null;
    this.#rejectAll(new WireProtocolError('wire connection closed'));
    this.#detachSocket();
  };

  #handleSocketError = (): void => {
    const error = new WireProtocolError('wire connection failed');
    this.#helloReject?.(error);
    this.#emitError(error);
  };

  async #decodeFrame(data: unknown): Promise<ServerFrame> {
    return ServerFrame.fromBinary(await binaryData(data));
  }

  #send(frame: ClientFrame): void {
    if (!this.#socket || this.#socket.readyState !== WEBSOCKET_OPEN) {
      throw new WireProtocolError('wire socket is not open');
    }
    this.#socket.send(frame.toBinary());
  }

  #rejectAll(error: WireProtocolError): void {
    for (const pending of this.#pendingRequests.values()) pending.reject(error);
    this.#pendingRequests.clear();
  }

  #emitError(error: WireProtocolError): void {
    for (const listener of this.#errorListeners) listener(error);
  }

  #detachSocket(): void {
    this.#socket?.removeEventListener('message', this.#handleMessage);
    this.#socket?.removeEventListener('close', this.#handleClose);
    this.#socket?.removeEventListener('error', this.#handleSocketError);
    this.#socket = null;
    this.#connectPromise = null;
  }

  #nextRequestId(): string {
    this.#requestSeq += 1;
    return `wire-request-${this.#requestSeq}`;
  }

  #nextFrameId(): string {
    this.#frameSeq += 1;
    return `wire-frame-${this.#frameSeq}`;
  }
}

async function binaryData(data: unknown): Promise<Uint8Array> {
  if (data instanceof Uint8Array) return data;
  if (data instanceof ArrayBuffer) return new Uint8Array(data);
  if (ArrayBuffer.isView(data)) {
    return new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
  }
  if (typeof Blob !== 'undefined' && data instanceof Blob) {
    return new Uint8Array(await data.arrayBuffer());
  }
  throw new WireProtocolError('wire frame data must be binary');
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function protobufBytes(message: BinaryMessage): Uint8Array<ArrayBuffer> {
  return new Uint8Array(message.toBinary());
}
