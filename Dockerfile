# Railway deploy image: official Chatto release + PORT/volume adapter.
FROM ghcr.io/chattocorp/chatto:0.4.2

COPY docker/railway-entrypoint.sh /usr/local/bin/railway-entrypoint.sh
RUN chmod +x /usr/local/bin/railway-entrypoint.sh

# Run as root so Railway volumes at /data are writable. Set RAILWAY_RUN_UID=0.
USER root

ENV PORT=8080 \
    CHATTO_WEBSERVER_PORT=8080 \
    CHATTO_NATS_EMBEDDED_ENABLED=true \
    CHATTO_NATS_EMBEDDED_DATA_DIR=/data \
    CHATTO_LIVEKIT_ENABLED=false \
    CHATTO_OPERATOR_API_ENABLED=false \
    CHATTO_LOG_FORMAT=json \
    CHATTO_LOG_LEVEL=info

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/railway-entrypoint.sh"]
