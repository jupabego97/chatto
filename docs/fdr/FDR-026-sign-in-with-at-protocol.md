# FDR-026: Sign in with AT Protocol

**Status:** Active
**Last reviewed:** 2026-05-19

## Overview

Users can sign in to Chatto using their AT Protocol identity — a DID issued by their PDS (typically `bsky.social`, but any compliant PDS works). The sign-in flow is offered alongside password and OIDC on the standard login page; once authentication completes, the resulting Chatto account behaves identically to one created by any other path. The intent is "unified identity": the same person signing in to two Chatto servers with the same ATProto handle is recognizable as the same person, without any federation between the servers.

## Behavior

- The login page shows a **Sign in with AT Protocol** form whenever the server has `[auth.atproto] enabled = true`. The form takes a single input — an ATProto handle (e.g. `alice.bsky.social`, `hendrik.mans.de`) — and a Continue button. A leading `@` is tolerated.
- Submitting the form redirects the browser to the user's PDS for the standard OAuth approval screen. After approval, the PDS redirects back to Chatto's callback endpoint.
- **First sign-in creates a Chatto account.** The user's DID is recorded as the stable identifier. Their handle is used to derive the Chatto login (so `alice.bsky.social` arrives as login `alice.bsky.social`). On collision with an existing Chatto login, a numeric suffix is appended (`-2`, `-3`, …).
- **First sign-in also seeds profile data.** Chatto reads the `app.bsky.actor.profile` record from the user's PDS and copies their display name and avatar into the new Chatto profile. The profile data is owned locally thereafter — Chatto does not re-sync from the PDS on later sign-ins.
- **Users without a Bluesky profile degrade gracefully.** A handle that has never used the `app.bsky.actor.profile` lexicon (e.g. a Frontpage-only user) signs in successfully; their Chatto display name defaults to their handle and they get the standard initial-letter avatar placeholder.
- **Subsequent sign-ins match by DID, not handle.** Handles can change at the protocol level; the DID is stable. The user's handle is re-resolved on each sign-in for display purposes but isn't used to identify them.
- **No email is collected.** The current scope set requests only `atproto`, which doesn't include access to the user's email. ATProto-only accounts therefore land in Chatto without a verified email — they can add one later through the standard email-management flow if they need one (e.g. for password reset, or to receive `owners.emails` privileges).
- **No ATProto credentials are retained.** Once the callback has identified the user, Chatto revokes the OAuth session immediately. Chatto does not store ATProto tokens and cannot act on the user's behalf against their PDS.
- **The resulting Chatto session is identical to any other.** Cookie + bearer token issued per FDR-023; from this point on the sign-in path is invisible to the rest of the application.

## Design Decisions

### 1. DID as the stable foreign key, handle as cached display

**Decision:** Store the user's DID as the primary link between an ATProto identity and a Chatto account. Re-resolve the handle on every sign-in for display purposes; don't persist it as an identity.
**Why:** Handles in ATProto can change (a user can move from `alice.bsky.social` to `alice.example.com` without losing their identity). The DID is the protocol's stable identifier. Persisting the handle as identity would silently break re-sign-in after a handle change.
**Tradeoff:** A handle-change between sign-ins means the user's Chatto login is stale until they update it through the normal rename flow. Acceptable — the alternative (renaming Chatto logins automatically on each sign-in) would be surprising and could cause its own confusion.

### 2. Handle-derived login with collision suffixing

**Decision:** On first sign-in, the Chatto login is derived from the ATProto handle directly (after lowercase + truncation to the existing 32-char login limit). On collision, suffix `-2`, `-3`, …, up to `-100`.
**Why:** Handles are already public, unique-within-PDS identifiers — perfect login material. Forcing the user to pick a fresh Chatto login at signup would be extra friction for no gain.
**Tradeoff:** A user whose preferred handle collides with an existing Chatto login lands with `-2` appended; they can fix it via the standard rename flow. Better than a forced disambiguation prompt at signup.

### 3. Phase 1 ships without the `account:email` scope

**Decision:** The initial implementation requests only the `atproto` scope. Email isn't collected; ATProto sign-ups land without a verified email.
**Why:** The `account:email` scope is a heavier consent ask — users see "this app wants to read your email address" on the approval screen. Phase 1 is identity-only; the email isn't load-bearing for anything yet. Adding the scope later is mechanical.
**Tradeoff:** ATProto users can't be auto-promoted to owner via `owners.emails`, and don't receive admin-driven email notifications, until they manually add an email to their Chatto account. The `account:email` scope can be added later when those use cases become real.

### 4. Profile mirroring on first sign-in only

**Decision:** Read `app.bsky.actor.profile` from the user's PDS once, on first sign-in, and copy display name + avatar into the Chatto profile. Don't re-sync on later sign-ins. Failures are logged and swallowed.
**Why:** Users expect their Chatto profile edits to stick — silently overwriting them on the next sign-in would be hostile. The first-sign-in copy gives ATProto users a head start (recognizable avatar, real name) instead of a blank-slate account; further changes are theirs to manage. Failure-as-silence handles users on PDSes without a bsky profile record (Frontpage, WhiteWind, …) without breaking sign-in.
**Tradeoff:** A user who changes their Bluesky display name won't see it propagate to Chatto. Considered correct — Chatto isn't a Bluesky mirror.

### 5. No ATProto tokens persist past sign-in

**Decision:** Revoke the OAuth session at the end of the callback. Store nothing.
**Why:** Phase 1 has no "act on the user's PDS" feature, so the tokens are pure attack surface. See ADR-032 for the general principle covering all external identity providers.
**Tradeoff:** Any future feature that needs to write to a user's PDS will have to do its own separate OAuth flow with its own scopes. Considered correct — those use cases deserve their own consent moment anyway.

### 6. ATProto sign-in only works at publicly-reachable or loopback URLs

**Decision:** The OAuth client uses `webserver.url` directly. When that URL is `127.0.0.1`/`localhost` (any port), the loopback dev-mode escape hatch kicks in; otherwise the deployment URL must be publicly reachable so the PDS can fetch the client metadata document. There is no third option.
**Why:** The constraint is baked into the ATProto OAuth spec. Encoding it directly in `webserver.url` matches how the rest of the codebase derives URLs and keeps the configuration surface minimal.
**Tradeoff:** Dev deployments on `*.orb.local`-style local-only hostnames can't use ATProto sign-in; operators have to either deploy publicly or develop against a loopback URL. The current `mise dev` setup uses loopback by default, so this is invisible to most local development.

## Related

- **ADRs:** ADR-032 (External Identity Integration Boundaries)
- **FDRs:** FDR-023 (Authentication & Sessions), FDR-022 (User Profile), FDR-018 (Account Lifecycle)

## Open Questions

- Whether to request `account:email` in a follow-up so ATProto sign-ups land with a verified email and become eligible for `owners.emails` auto-promotion. The trade-off is the heavier consent screen.
- Whether to define a `run.chatto.*` lexicon for Chatto-specific profile data on the user's PDS (cross-server discovery, "find me on these other Chatto servers"). Out of scope for Phase 1.
- Account linking: today, a user with both a Chatto password account and an ATProto handle that resolves to the same person ends up with two separate Chatto accounts. A future settings-page action could merge them on explicit consent.
