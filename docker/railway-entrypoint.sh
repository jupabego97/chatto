#!/bin/sh
# Railway adapter: map $PORT, prepare /data, then hand off to the official
# image entrypoint (PUID/PGID remapping + nats CLI context).
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

# Fail fast with a clear message when required secrets are missing.
missing=""
for var in \
  CHATTO_WEBSERVER_COOKIE_SIGNING_SECRET \
  CHATTO_CORE_SECRET_KEY \
  CHATTO_CORE_ASSETS_SIGNING_SECRET
do
  eval "val=\${$var:-}"
  if [ -z "$val" ]; then
    missing="$missing $var"
  fi
done
if [ -n "$missing" ]; then
  echo "railway-entrypoint: missing required env vars:$missing" >&2
  echo "railway-entrypoint: copy values from .env.railway.example into Railway Variables" >&2
  exit 1
fi

# Railway volumes are root-owned. The official entrypoint drops to PUID:PGID,
# so chown /data (and /config) before that handoff.
mkdir -p /data /config
if [ "$(id -u)" = "0" ]; then
  chown "${PUID:-1000}:${PGID:-1000}" /data /config
else
  echo "railway-entrypoint: warning: not running as root (uid=$(id -u)); set RAILWAY_RUN_UID=0 so /data can be prepared" >&2
fi

# Probe writability early — embedded NATS will otherwise fail obscurely.
data_dir="${CHATTO_NATS_EMBEDDED_DATA_DIR}"
if ! touch "${data_dir}/.chatto-write-test" 2>/dev/null; then
  echo "railway-entrypoint: cannot write to ${data_dir}" >&2
  echo "railway-entrypoint: attach a Railway Volume at /data and set RAILWAY_RUN_UID=0" >&2
  exit 1
fi
rm -f "${data_dir}/.chatto-write-test"

echo "railway-entrypoint: starting chatto on port ${CHATTO_WEBSERVER_PORT} (data=${CHATTO_NATS_EMBEDDED_DATA_DIR})"

exec /usr/local/bin/docker-entrypoint.sh "$@"
