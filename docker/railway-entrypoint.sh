#!/bin/sh
# Railway adapter for the official Chatto image.
# - Maps Railway $PORT
# - Prepares a writable data directory
# - Runs /chatto directly (skip su-exec privilege drop; use RAILWAY_RUN_UID=0)
set -eu

echo "railway-entrypoint: boot uid=$(id -u) gid=$(id -g) PORT=${PORT:-} RAILWAY_VOLUME_MOUNT_PATH=${RAILWAY_VOLUME_MOUNT_PATH:-}"

# Prefer Railway's injected PORT; fall back to 8080 for healthchecks/networking.
if [ -z "${PORT:-}" ]; then
  export PORT=8080
  echo "railway-entrypoint: PORT was empty; defaulting to 8080"
fi
export CHATTO_WEBSERVER_PORT="$PORT"

# Single-node defaults.
export CHATTO_NATS_EMBEDDED_ENABLED="${CHATTO_NATS_EMBEDDED_ENABLED:-true}"
export CHATTO_LIVEKIT_ENABLED="${CHATTO_LIVEKIT_ENABLED:-false}"
export CHATTO_LOG_FORMAT="${CHATTO_LOG_FORMAT:-json}"
export CHATTO_LOG_LEVEL="${CHATTO_LOG_LEVEL:-info}"
export CHATTO_OPERATOR_API_ENABLED="${CHATTO_OPERATOR_API_ENABLED:-false}"

# Pick a writable data directory.
# 1) Railway volume mount path, if present
# 2) /data (when a volume is attached there)
# 3) /tmp/chatto-data (ephemeral fallback so the process can still start)
data_dir="${CHATTO_NATS_EMBEDDED_DATA_DIR:-}"
if [ -z "$data_dir" ] && [ -n "${RAILWAY_VOLUME_MOUNT_PATH:-}" ]; then
  data_dir="$RAILWAY_VOLUME_MOUNT_PATH"
fi
if [ -z "$data_dir" ]; then
  data_dir=/data
fi

mkdir -p "$data_dir" 2>/dev/null || true
if ! touch "$data_dir/.chatto-write-test" 2>/dev/null; then
  echo "railway-entrypoint: cannot write to $data_dir; falling back to /tmp/chatto-data" >&2
  data_dir=/tmp/chatto-data
  mkdir -p "$data_dir"
fi
rm -f "$data_dir/.chatto-write-test" 2>/dev/null || true
export CHATTO_NATS_EMBEDDED_DATA_DIR="$data_dir"

# Fail fast when required secrets are missing.
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
  echo "railway-entrypoint: add them in Railway → Variables (see .env.railway.example)" >&2
  exit 1
fi

if [ -z "${CHATTO_WEBSERVER_URL:-}" ]; then
  if [ -n "${RAILWAY_PUBLIC_DOMAIN:-}" ]; then
    export CHATTO_WEBSERVER_URL="https://${RAILWAY_PUBLIC_DOMAIN}"
    echo "railway-entrypoint: derived CHATTO_WEBSERVER_URL=$CHATTO_WEBSERVER_URL"
  else
    echo "railway-entrypoint: warning: CHATTO_WEBSERVER_URL is unset (Generate Domain + set the variable)" >&2
  fi
fi

echo "railway-entrypoint: starting /chatto on :$CHATTO_WEBSERVER_PORT data=$CHATTO_NATS_EMBEDDED_DATA_DIR url=${CHATTO_WEBSERVER_URL:-}"

# Run the binary directly. Config file is optional; env vars supply production config.
exec /chatto start -c /config/chatto.toml
