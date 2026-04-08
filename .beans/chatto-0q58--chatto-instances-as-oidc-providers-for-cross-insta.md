---
# chatto-0q58
title: Chatto instances as OIDC providers for cross-instance SSO
status: draft
type: feature
priority: normal
created_at: 2026-03-20T16:30:52Z
updated_at: 2026-03-20T16:30:52Z
---

## Motivation

Currently, when adding a remote instance, users type their password into a form on the local client that sends it directly to the remote server via `POST /auth/login`. This works but has trust and security concerns:

- User credentials are typed into a UI served by a different origin than the server they authenticate against
- Bearer tokens are stored in localStorage (XSS-accessible)
- No standard protocol — custom auth flow that only Chatto understands

## Vision

Each Chatto instance acts as an **OIDC provider** (identity provider). When a user wants to connect to Instance B from Instance A's client (or from a standalone client):

1. Client redirects the user to Instance B's `/authorize` endpoint
2. User authenticates on Instance B's own login page (same-origin, cookie auth — secure)
3. Instance B redirects back to the client with an authorization code
4. Client exchanges the code for an access token + refresh token
5. Client uses the access token for GraphQL API calls

The user's credentials **never leave Instance B's origin**. The client only ever receives tokens.

## What Chatto Already Has

- User accounts with password + OAuth login
- Session management (cookie-based)
- Token issuance (cht_AT opaque tokens with NATS KV TTL)
- Sliding window token refresh (re-put on validation)
- /api/instance discovery endpoint with wildcard CORS

## What Needs to Be Built

### Backend (OIDC Provider)

- /.well-known/openid-configuration — Discovery document
- /.well-known/jwks.json — Public keys for token verification
- /oauth/authorize — Authorization endpoint with consent screen
- /oauth/token — Token endpoint (authorization code grant)
- ID token issuance (JWT with sub, email, preferred_username, picture claims)
- Refresh token support (separate from access tokens)
- Client registration — either:
  - **Static**: Well-known client ID for the Chatto web client (simplest)
  - **Dynamic**: RFC 7591 dynamic client registration (future, for third-party apps)

**Go libraries to evaluate:**
- github.com/zitadel/oidc — Full OIDC provider/relying party implementation
- github.com/ory/fosite — OAuth2/OIDC framework (more low-level)
- github.com/coreos/go-oidc — Relying party only (not provider)

### Backend (OIDC Client / Relying Party)

When Instance A's backend needs to verify tokens from Instance B (e.g., for server-to-server federation in the future):
- Standard OIDC token validation via JWKS
- User mapping (create local shadow user from ID token claims)

### Frontend

- Replace the password form in /chat/instances/add/[hostname] with an OAuth redirect flow:
  1. Probe /api/instance (already done) — get OIDC discovery URL
  2. Redirect to {instance}/.well-known/openid-configuration — find authorize endpoint
  3. Initiate Authorization Code + PKCE flow
  4. Handle redirect callback with authorization code
  5. Exchange code for tokens
  6. Store access token + refresh token in localStorage
- urql authExchange for automatic token refresh on 401
- Token rotation: proactive refresh before expiry (e.g., when token has <5 min left)

### Consent Screen

A new page on each instance: "The Chatto client at {origin} wants to access your account"
- Show what scopes are requested (openid, profile)
- "Allow" / "Deny" buttons
- Remember consent for known client IDs (don't re-prompt every time)

## Scope Definitions

Start minimal:
- openid — Required. Returns sub claim (user ID)
- profile — Returns preferred_username, name, picture
- spaces:read — List and read spaces/rooms the user is a member of
- messages:read — Read messages in joined rooms
- messages:write — Post messages

Future:
- admin — Instance admin operations
- offline_access — Refresh token grant

## Migration Path

1. **Phase 1**: Add OIDC provider endpoints to Chatto. Keep password-based flow as fallback.
2. **Phase 2**: Update /api/instance to advertise OIDC support. Frontend prefers OIDC when available, falls back to password.
3. **Phase 3**: Deprecate password-based cross-origin flow. Show warning when connecting to instances that don't support OIDC.
4. **Phase 4**: Remove password flow entirely.

## Trade-offs

**Pros:**
- Credentials never leave the provider's origin — much better security posture
- Standard protocol — third-party apps can authenticate too
- Token refresh is spec-defined (no custom sliding window needed for cross-origin)
- SSO across all connected Chatto instances
- Foundation for federation (Instance A trusting Instance B's users)

**Cons:**
- Significantly more backend work than the current password flow
- Redirect-based flow is more complex UX (leaves the page, comes back)
- Need to handle PKCE, state params, nonce validation correctly
- Client registration adds operational complexity
- Mobile/desktop apps need to handle redirect via system browser or custom URI scheme

**Open questions:**
- Should we use an off-the-shelf OIDC library or implement a minimal subset?
- How do we handle the consent screen for the well-known Chatto client? (Auto-approve?)
- Should access tokens be JWTs (verifiable without network call) or opaque (like current cht_AT)?
- How does this interact with instances that already use external OIDC providers (Google, GitHub)? Chain of trust?
