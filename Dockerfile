# Railway deploy image: official Chatto release + PORT/volume adapter.
# Pin the tag for reproducible deploys; bump when you want a newer release.
FROM ghcr.io/chattocorp/chatto:v0.4.2

COPY docker/railway-entrypoint.sh /usr/local/bin/railway-entrypoint.sh
RUN chmod +x /usr/local/bin/railway-entrypoint.sh

# Volumes on Railway are root-owned; the official entrypoint remaps via PUID/PGID
# when started as root. Prefer RAILWAY_RUN_UID=0 in Railway variables.
USER root

ENV CHATTO_NATS_EMBEDDED_ENABLED=true \
    CHATTO_NATS_EMBEDDED_DATA_DIR=/data \
    CHATTO_WEBSERVER_PORT=4000 \
    CHATTO_LIVEKIT_ENABLED=false \
    CHATTO_LOG_FORMAT=json \
    CHATTO_LOG_LEVEL=info \
    PUID=1000 \
    PGID=1000

# Persistent data: attach a Railway Volume at /data (not VOLUME here — Railway rejects it).
EXPOSE 4000

ENTRYPOINT ["/usr/local/bin/railway-entrypoint.sh"]
CMD ["start", "-c", "/config/chatto.toml"]
