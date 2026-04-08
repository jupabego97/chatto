---
# chatto-aoke
title: OAuth Authorization Code + PKCE flow for instance auth
status: draft
type: feature
priority: normal
created_at: 2026-03-24T17:21:05Z
updated_at: 2026-03-24T17:21:37Z
parent: chatto-wadw
---

Replace direct credential entry with an OAuth 2.0 Authorization Code + PKCE flow for authenticating against remote (and all non-origin) Chatto instances. Each instance acts as an OAuth authorization server. Clients (origin instances, static SPAs, desktop apps) use the standard public client PKCE flow — redirect to remote's login UI, get a code, exchange for a bearer token.

## Motivation

The current add-instance flow requires users to type their credentials for Server B into a UI served by Server A (or a standalone client). This has several problems:

1. **Only password auth works** — OAuth providers (Google, GitHub) can't be used because they require server-side callbacks on the remote instance
2. **No registration** — new users can't sign up on a remote instance from within the client
3. **Trust concern** — credentials for one server are entered into another server's UI
4. **All auth methods must be re-implemented** — every new auth method the remote supports needs a separate cross-origin implementation in the client

With OAuth Authorization Code + PKCE, the remote instance shows its own login/signup UI with all its supported auth methods, and the client just receives a code to exchange for a token.

## Client Types

The flow must work for all three client scenarios:

| Client | Example | redirect_uri |
|--------|---------|-------------|
| Origin instance | SPA served by a Chatto backend | `https://origin.example.com/chat/instances/callback` |
| Static SPA | CDN-hosted, no backend | `https://app.chatto.chat/chat/instances/callback` |
| Desktop app | Tauri/Electron | `http://localhost:{port}/callback` or custom scheme |

## Flow

```
Client                          Remote Instance
──────                          ───────────────
1. Open browser/popup ────────► GET /oauth/authorize
   ?response_type=code              ?redirect_uri=...
   &code_challenge=SHA256(verifier) &state=<random>
   &code_challenge_method=S256

                                2. Show login UI
                                   (password, Google, GitHub, etc.)
                                   User authenticates normally.

                                3. Generate authorization code
                                   Store in AUTH_TOKENS KV:
                                     key: grant:<code>
                                     val: {user_id, redirect_uri,
                                           code_challenge, method}
                                     TTL: 5 minutes (per-key via Nats-TTL)

4. Receive redirect ◄────────── 302 redirect_uri?code=<code>&state=<state>

5. Exchange code ─────────────► POST /oauth/token
   {grant_type: authorization_code,  
    code: <code>,
    code_verifier: <verifier>,
    redirect_uri: <same as step 1>}

                                6. Validate:
                                   - code exists and not expired
                                   - SHA256(code_verifier) == code_challenge
                                   - redirect_uri matches
                                   Delete code (single-use).
                                   Create bearer token via CreateAuthToken().

7. Receive token ◄──────────── {access_token: "cht_AT...",
                                 token_type: "Bearer",
                                 user: {id, login, displayName, ...}}
```

## Design Decisions

### No client registration
Public clients only (no client_secret). Any client presenting a valid PKCE challenge is accepted. The redirect_uri is the sole identifier — validated against policy (see below).

### Redirect URI validation
- Web clients: must be HTTPS (except localhost for development)
- Desktop apps: `http://localhost:*` or custom `chatto://` scheme
- Configurable allowlist in `chatto.toml` for additional origins (optional)
- Default policy: accept any HTTPS origin + localhost (permissive, since PKCE provides the security)

### Storage: AUTH_TOKENS KV with key prefix
Authorization codes stored in the existing `AUTH_TOKENS` bucket with `grant:` key prefix. Per-key TTL of 5 minutes via NATS `Nats-TTL` header / `KeyTTL` option on `Create()`. No new KV bucket needed.

### No library — DIY implementation
The Authorization Code + PKCE flow is ~200 lines of Go using stdlib (`crypto/rand`, `crypto/sha256`, `encoding/base64`). Libraries like fosite/osin add more complexity than they remove for this focused use case.

### Relation to existing OAuth callbacks
The existing `/auth/:provider/callback` flow (Google, GitHub) continues to work for the origin instance's own login. The new `/oauth/authorize` endpoint wraps the entire auth experience — whichever method the user picks on the remote, the result is an authorization code redirected back to the client.

## Child Tasks

See child beans for implementation breakdown.
