# ADR-004: Authorization Enforced at the API Boundary, Not in Core

**Date:** 2026-03-01

## Context

Chatto has two API surfaces that invoke business logic:

1. **GraphQL** — user-facing, untrusted callers (browser clients)
2. **NATS request-reply** — internal/extension API, trusted callers (background jobs, NATS handlers, webhooks)

A common pattern is to embed authorization checks inside every business logic function. This ensures safety but adds redundant checks when called from trusted contexts, makes core functions harder to test (every test needs auth setup), and couples authorization policy to business logic.

## Decision

Enforce authorization at the **API boundary** (GraphQL resolvers), not in core business logic:

- **GraphQL resolvers** call `Can*` permission-check functions and `require*` auth helpers before invoking core methods. If the caller lacks permission, the resolver returns an error without calling core.
- **Core functions** assume the caller is authorized. They document their authorization requirements in comments (e.g., "Caller must verify CanCreateRoom before calling") but do not check permissions themselves.
- **NATS handlers** call core directly since they operate in a trusted context.

## Consequences

- **Core is reusable**: Core functions can be called from any trusted context without redundant permission checks. Background jobs, webhook handlers, and NATS services call core directly.
- **No double-checking**: A GraphQL resolver that checks `CanPostMessage` doesn't trigger another permission check inside `PostMessage`. One check, at the boundary.
- **Testability**: Core unit tests focus on business logic without needing to set up permission fixtures. Authorization is tested separately at the resolver level.
- **Clear responsibility**: Reading a resolver tells you exactly what permissions are required. Reading a core function tells you what it does. Concerns are separated.
- **Risk of forgetting checks**: A new GraphQL resolver that calls core without checking permissions would bypass authorization. Mitigated by code review, the `require*` helper convention, and the authorization reference table in `authorization.md`.
- **Audit logging can be added orthogonally**: Since authorization happens at the boundary, audit logging doesn't need to be coupled to business logic.
