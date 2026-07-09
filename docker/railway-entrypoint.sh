#!/bin/sh
# Railway adapter: map $PORT and apply safe single-node defaults, then hand off
# to the official image entrypoint (PUID/PGID + nats CLI context).
set -eu

if [ -n "${PORT:-}" ]; then
  export CHATTO_WEBSERVER_PORT="$PORT"
fi

# Single-node Railway deploy: embedded NATS + persistent /data volume.
export CHATTO_NATS_EMBEDDED_ENABLED="${CHATTO_NATS_EMBEDDED_ENABLED:-true}"
export CHATTO_NATS_EMBEDDED_DATA_DIR="${CHATTO_NATS_EMBEDDED_DATA_DIR:-/data}"

# LiveKit needs UDP media ports that Railway does not expose; keep calls off
# unless the operator explicitly configures an external LiveKit cluster.
export CHATTO_LIVEKIT_ENABLED="${CHATTO_LIVEKIT_ENABLED:-false}"

export CHATTO_LOG_FORMAT="${CHATTO_LOG_FORMAT:-json}"
export CHATTO_LOG_LEVEL="${CHATTO_LOG_LEVEL:-info}"

exec /usr/local/bin/docker-entrypoint.sh "$@"


