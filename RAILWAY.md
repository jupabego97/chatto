# Railway deployment for Chatto

Single-node setup using the official image (`ghcr.io/chattocorp/chatto`) with **embedded NATS** and a persistent volume at `/data`.

Voice/video calls (LiveKit) stay **disabled** by default: Railway does not expose the UDP media ports LiveKit needs. Chat, DMs, files, and presence work without LiveKit.

## 1. Deploy from GitHub

1. Open [Railway](https://railway.app) → **New Project** → **Deploy from GitHub repo**.
2. Select `jupabego97/chatto` (or your fork).
3. Railway should detect `Dockerfile` + `railway.toml` and build automatically.

## 2. Attach a Railway volume (required)

Do **not** use a Dockerfile `VOLUME` instruction — Railway rejects it. Persist data with a Railway Volume instead.

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
| `PORT` | `8080` |
| `CHATTO_WEBSERVER_URL` | `https://${{RAILWAY_PUBLIC_DOMAIN}}` |
| `CHATTO_OWNERS_EMAILS` | your admin email |
| `CHATTO_WEBSERVER_COOKIE_SIGNING_SECRET` | 64 hex chars |
| `CHATTO_WEBSERVER_COOKIE_ENCRYPTION_SECRET` | 64 hex chars |
| `CHATTO_CORE_SECRET_KEY` | 64 hex chars |
| `CHATTO_CORE_ASSETS_SIGNING_SECRET` | 64 hex chars |

Also set the public domain **target port** to `8080` in Networking if Railway asks for one.

Healthchecks are off by default in `railway.toml` until deploy logs show a clean boot. After that, set healthcheck path to `/healthz`.

Generate secrets in PowerShell:

```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Maximum 256) })
```

Full template: [`.env.railway.example`](.env.railway.example).

`PORT` is injected by Railway; the entrypoint maps it to `CHATTO_WEBSERVER_PORT`.

## 5. Create the first admin account

Registration from the web UI needs SMTP (email verification codes). Without SMTP, create the owner with the Operator CLI inside the running container.

### Option A — Operator CLI (no SMTP)

1. Make sure the service is running and `CHATTO_OPERATOR_API_ENABLED=true` (default in this fork).
2. Open a shell in the container:
   - Railway dashboard → service → **Shell**, or
   - `railway ssh` from your machine (after `railway link`).
3. Create the owner (change login/email/password):

```sh
/chatto operator user create \
  --login admin \
  --display-name "Admin" \
  --password 'CambiaEstaClave123!' \
  --verified-email tu@email.com \
  --role owner
```

4. Open the public URL and sign in with `admin` / that password.

List users later with:

```sh
/chatto operator user list
```

### Option B — Web register + Resend SMTP

1. Create an API key at [resend.com](https://resend.com).
2. Set Railway variables:

```env
CHATTO_SMTP_ENABLED=true
CHATTO_SMTP_HOST=smtp.resend.com
CHATTO_SMTP_PORT=465
CHATTO_SMTP_TLS=implicit
CHATTO_SMTP_USERNAME=resend
CHATTO_SMTP_PASSWORD=re_xxxxxxxx
CHATTO_SMTP_FROM=onboarding@resend.dev
CHATTO_OWNERS_EMAILS=tu@email.com
```

3. Redeploy, open the public URL → **Register** with that email → enter the 6-digit code.
4. That account becomes owner if the email matches `CHATTO_OWNERS_EMAILS`.

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

1. Bump the `FROM ghcr.io/chattocorp/chatto:X.Y.Z` tag in `Dockerfile` (no `v` prefix; e.g. `0.4.2`).
2. Push to GitHub; Railway redeploys.
3. Keep the `/data` volume attached so JetStream data survives.

## License note

Chatto is AGPL-3.0-or-later. Self-hosting a modified version on Railway still requires complying with that license.
