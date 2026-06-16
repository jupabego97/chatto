import { WireClient, WireProtocolError } from '$lib/wire';
import type { StreamEvent } from '$lib/pb/chatto/wire/v1/protocol_pb';
import type { GraphQLClient } from './graphqlClient.svelte';

export interface WireEventLogger {
  stop(): void;
}

export function startWireEventLogger(serverId: string, gqlClient: GraphQLClient): WireEventLogger {
  const client = new WireClient({
    url: gqlClient.wireUrl,
    token: gqlClient.token
  });

  const stopEventListener = client.onEvent((event) => {
    console.log(`[wire:${serverId}] event`, wireEventLogPayload(event));
  });

  const stopErrorListener = client.onError((error) => {
    console.error(`[wire:${serverId}] error`, wireErrorLogPayload(error));
  });

  try {
    void client
      .connect()
      .then((hello) => {
        console.log(`[wire:${serverId}] connected`, {
          protocolVersion: hello.protocolVersion,
          serverVersion: hello.serverVersion,
          methods: hello.methods,
          features: hello.features
        });
      })
      .catch((error: unknown) => {
        if (client.status === 'closed') return;
        console.error(`[wire:${serverId}] connect failed`, wireErrorLogPayload(error));
      });
  } catch (error: unknown) {
    console.error(`[wire:${serverId}] connect failed`, wireErrorLogPayload(error));
  }

  return {
    stop() {
      stopEventListener();
      stopErrorListener();
      client.dispose();
    }
  };
}

function wireEventLogPayload(event: StreamEvent) {
  return {
    eventId: event.eventId,
    eventType: event.eventType,
    deliveryCursor: cursorDebug(event.deliveryCursor),
    payload: event.payload.case ?? null,
    invalidates: event.invalidates.map((hint) => ({
      kind: hint.kind,
      id: hint.id
    })),
    event
  };
}

function wireErrorLogPayload(error: unknown) {
  if (error instanceof WireProtocolError) {
    return {
      message: error.message,
      code: error.wireError?.code,
      requestId: error.wireError?.requestId,
      retryable: error.wireError?.retryable
    };
  }
  if (error instanceof Error) {
    return { message: error.message };
  }
  return error;
}

function cursorDebug(cursor: string) {
  if (!cursor) return { present: false };
  return {
    present: true,
    length: cursor.length,
    suffix: cursor.slice(-8)
  };
}
