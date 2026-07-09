# Railway deployment for Chatto

Single-node setup using the official image (`ghcr.io/chattocorp/chatto`) with **embedded NATS** and a persistent volume at `/data`.

Voice/video calls (LiveKit) stay **disabled** by default: Railway does not expose the UDP media ports LiveKit needs. Chat, DMs, files, and presence work without LiveKit.

## 1. Deploy from GitHub

1. Open [Railway](https://railway.app) → **New Project** → **Deploy from GitHub repo**.
2. Select `jupabego97/chatto` (or your fork).
3. Railway should detect `Dockerfile` + `railway.toml` and build automatically.

## 2. Attach a volume (required)

Without a volume, all chat data is lost on every redeploy.

1. Right-click the canvas → **Volume** (or Command Palette).
2. Attach it to the Chatto service.
3. Mount path: **`/data`**

## 3. Generate a public domain

In the service → **Settings** → **Networking** → **Generate Domain**.

Copy the URL (e.g. `https://chatto-production-xxxx.up.railway.app`).

## 4. Set environment variables

In the service → **Variables**, add at least:

| Variable | Value |
| --- | --- |
| `RAILWAY_RUN_UID` | `0` |
| `CHATTO_WEBSERVER_URL` | `https://${{RAILWAY_PUBLIC_DOMAIN}}` |
| `CHATTO_OWNERS_EMAILS` | your admin email |
| `CHATTO_WEBSERVER_COOKIE_SIGNING_SECRET` | 64 hex chars |
| `CHATTO_WEBSERVER_COOKIE_ENCRYPTION_SECRET` | 64 hex chars |
| `CHATTO_CORE_SECRET_KEY` | 64 hex chars |
| `CHATTO_CORE_ASSETS_SIGNING_SECRET` | 64 hex chars |
| SMTP vars | see `.env.railway.example` |

Generate secrets in PowerShell:

```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Maximum 256) })
```

Full template: [`.env.railway.example`](.env.railway.example).

`PORT` is injected by Railway; the entrypoint maps it to `CHATTO_WEBSERVER_PORT`.

## 5. First login

1. Open the public URL.
2. Register with the email listed in `CHATTO_OWNERS_EMAILS`.
3. Complete email verification (SMTP must work).
4. That account becomes an owner.

## Files in this repo

| File | Role |
| --- | --- |
| `Dockerfile` | Thin wrapper around the official release image |
| `docker/railway-entrypoint.sh` | Maps `$PORT` and Railway defaults |
| `railway.toml` | Builder + `/healthz` healthcheck |
| `.env.railway.example` | Variable checklist |

## Limits on Railway

- **No LiveKit / calls** unless you run LiveKit elsewhere and set `CHATTO_LIVEKIT_*`.
- **Single replica** with embedded NATS (do not scale horizontally with this setup).
- Prefer **S3-compatible storage** for heavy file uploads (`CHATTO_CORE_ASSETS_STORAGE_BACKEND=s3` + S3 vars).
- Pin the image tag in `Dockerfile` when you want controlled upgrades.

## Upgrade

1. Bump the `FROM ghcr.io/chattocorp/chatto:vX.Y.Z` tag in `Dockerfile`.
2. Push to GitHub; Railway redeploys.
3. Keep the `/data` volume attached so JetStream data survives.

## License note

Chatto is AGPL-3.0-or-later. Self-hosting a modified version on Railway still requires complying with that license.
