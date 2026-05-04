# ADR-027: User Identity and OIDC Binding for the Server Model

**Date:** 2026-05-04

## Context

[ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) collapses the Instance + Space layers into a single Server. That collapse changes nothing about *how* users authenticate or what their identifiers look like, but it does force a clear restatement of identity semantics — because once "Instance" stops being a concept, phrases like "instance-scoped user ID" or "the user's instance email" need to be rephrased without inheriting the old terminology.

Identity in Chatto has historically combined three things in a single user record: a per-instance opaque user ID (NanoID, see [ADR-022](ADR-022-nanoid-with-entity-prefixes.md)), a set of human-meaningful identifiers (login, verified emails), and zero or more authentication credentials (a built-in password hash and/or an OIDC `sub` from an external identity provider). Email addresses change over time — people switch jobs, change names, leave providers — so binding cross-IdP identity to email is wrong: a user re-signing-in with the same OIDC provider but a new email must end up in the same Chatto user record.

A separate, related question is what happens *across* servers. A user logged into both `chat.acme.com` and `chat.contoso.com` is, from the user's perspective, the same person — but the two servers have no shared infrastructure, no shared user table, and no agreed authority on "same person." The new model needs to say something explicit about this so that future work has a clear constraint to satisfy.

## Decision

### Stable internal user IDs

Each server has its own user records. Each user record is identified by a 14-character NanoID with the `U` prefix, exactly as today (per [ADR-022](ADR-022-nanoid-with-entity-prefixes.md)). User IDs are opaque, server-local, and never reused. They are the stable referent for everything else in the system: messages, role assignments, room memberships, audit logs.

User IDs do **not** encode the server they belong to. A `U…` ID is meaningful only inside the server that issued it.

### OIDC `sub` is the binding key for external identity

When a user authenticates via OIDC, the `sub` claim from the IdP is the binding key for matching the OIDC identity to a Chatto user record. **Email is never used as a binding key.** Specifically:

- On first OIDC sign-in, a new user record is created and `oidc_sub` is stored on it.
- On subsequent OIDC sign-ins, the user record is looked up by `(issuer, sub)` and updated. The current email from the OIDC userinfo response is recorded as a verified email, but is not used to match the record.
- If an OIDC user later changes their email at the IdP, they sign back in with the same `(issuer, sub)` and continue using the same Chatto user record. Their email on the record updates.
- Multiple IdPs are supported by storing `(issuer, sub)` pairs, not just `sub`. Two providers can legitimately share the same `sub` value.

### Built-in email-password auth and OIDC coexist on the same record

The user record holds an optional password hash and an optional set of `(issuer, sub)` OIDC bindings. Either, both, or neither may be present:

- **Built-in only**: classic email-password sign-up, no OIDC.
- **OIDC only**: sign-in via Hub or another IdP, no password.
- **Both**: a user who started on built-in auth and later linked an OIDC provider (or vice versa). They can sign in either way; both paths land on the same user record.

Linking and unlinking are explicit user actions (separate ADR if and when the linking UX is built); the data model already supports it.

### `owners.emails` config remains, with unchanged semantics

The `owners.emails` config block continues to designate server-level owners by email match against the user's *verified* emails (only). Unverified or pending emails are ignored. The matching is implicitly per-server, since one process is one server.

The owner short-circuit in the permission resolver (see [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md), [ADR-028](ADR-028-permission-model-post-merge.md)) reads exactly the same as today: any verified email matching `owners.emails` grants owner-level access, bypassing the role hierarchy.

### Cross-server identity consolidation is out of scope

The server model deliberately makes no attempt to consolidate "the same person on different servers." Two servers each have their own user table, their own user IDs, and their own opinion of who someone is. A user signed into both Acme and Contoso is, to the system, two distinct user records that happen to share an email or an OIDC `sub` at the IdP layer.

If cross-server "this is the same person" consolidation is ever needed, it belongs in the *client* (or in Hub, as a directory), not in any individual server's identity store. This is tracked separately and is out of scope here.

### IdP-agnosticism

A Chatto server doesn't need to know whether its IdP is Chatto Hub, an external OIDC provider, or both. The OIDC handshake is standard; the server stores `(issuer, sub)` pairs. Hub acts as the IdP for B2C users; B2B operators can configure a different IdP (their corporate SSO, for example) or bridge through Hub. The server treats them identically.

## Consequences

- **Email changes are safe**. A user can change the email they sign in with (at the IdP or on the built-in account) without losing their identity, history, or role assignments. The server's stable referent is the user ID, bound by `(issuer, sub)` for OIDC and by the email-verified login for built-in auth.
- **`owners.emails` keeps working**. No migration needed for existing config files. The semantics — verified emails only — are explicitly carried over.
- **Cross-server identity isn't built into the server**. Multi-server users today see themselves as multiple accounts; that's the explicit posture of the new model. A future Hub or client-side consolidation is unconstrained by server-level decisions.
- **Linked accounts are a future feature, not a present one**. The data model supports a user record having both a password and one or more OIDC bindings. The UX to manage that linking is not part of this ADR.
- **Anonymous users are still server-scoped**. Anonymous browse access (per [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) — anonymous callers can see the server's public surface) doesn't require identity binding; it's a separate code path.
- **Migration impact**: existing user records carry forward unchanged. The user ID, password hash, OIDC `sub`, and verified emails are all already present in the user record — there is no schema change required by this ADR. The new clarity is in *how* we describe and reason about identity, not in how it's stored.

## References

- Closes [#290](https://github.com/chattocorp/chatto/issues/290).
- Builds on [ADR-022](ADR-022-nanoid-with-entity-prefixes.md) (NanoID format).
- Companion to [ADR-021](ADR-021-consolidate-instance-and-space-into-server.md) (Server consolidation), [ADR-028](ADR-028-permission-model-post-merge.md) (permission model), [ADR-029](ADR-029-server-destination-schema.md) (destination schema), [ADR-030](ADR-030-dm-as-room.md) (DM-as-room).
