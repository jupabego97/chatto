---
name: chatto-features
description: Audit and update feature documentation in agent-docs/features/. Verifies each feature doc against the actual codebase, flags discrepancies, and can create new feature docs for undocumented features.
---

# Feature Documentation Audit & Update

Audit the feature docs in `agent-docs/features/` against the actual codebase, flag discrepancies, and update or create docs as needed.

## What These Docs Are

The `agent-docs/features/` directory contains **high-level feature descriptions** that act as guardrails for agents. They explain how features are designed and how their parts fit together from a user perspective, **without going into implementation details**.

These docs should read like a product spec, not a code walkthrough. They answer:

- What does the feature do?
- How does it behave from the user's perspective?
- What permissions gate it?
- What are the key design decisions and non-obvious interactions?

### What These Docs Are NOT

- NOT code documentation — no function signatures, no proto field tags, no GraphQL schema dumps
- NOT file indexes — no "Key Files" tables listing every relevant source file
- NOT implementation guides — no KV key patterns, no NATS subject formats, no dataloader details

An agent can always `grep` for those things. The feature docs exist to provide the **context and design intent** that can't be derived from reading code alone.

## Writing Style

Study the existing docs (especially `replies-and-threads.md`) for the right tone and level of detail:

- **Use bullet points and short paragraphs**, not code blocks
- **Describe behavior**, not implementation ("echo events appear in the room timeline" not "echo events are published to `space.{s}.room.{r}.msg.{echoEventId}`")
- **Mention permission strings** (e.g., `message.reply`) because those are part of the feature design, but don't describe how permission checks are implemented
- **Mention key design decisions** that aren't obvious from the code (e.g., "reactions attach to event IDs, not body IDs, so echoes have independent reactions")
- **Omit sections that don't add value** — a simple feature might just need "Overview" and "Permissions"
- **No code snippets** — if you're writing GraphQL schema, proto definitions, or Go/TypeScript code, you're going too deep

## How To Run

### Mode 1: Audit All Feature Docs (default)

When invoked without arguments, audit every doc in `agent-docs/features/`.

### Mode 2: Audit a Specific Feature

When invoked with a feature name (e.g., `/chatto-features thread-reply-echo`), audit only that feature doc.

### Mode 3: Create a New Feature Doc

When invoked with `new <feature-name>` (e.g., `/chatto-features new reactions`), create a new feature doc by examining the codebase.

## Audit Process

For each feature doc being audited, launch a **dedicated Explore subagent** that:

1. **Reads the feature doc** in full
2. **Verifies each claim** against the codebase:
   - Permission names — do the constants exist in `cli/internal/core/permissions.go`?
   - Described behaviors — does the code actually work this way?
   - Design claims — are the stated interactions accurate?
3. **Returns a structured report** with:
   - **Verified**: Claims confirmed by code
   - **Discrepancies**: Claims that don't match the code (with details)
   - **Missing from doc**: Significant user-facing behavior not mentioned
   - **Stale claims**: Behaviors or interactions that no longer exist

**Launch audits for multiple features in parallel** using the Agent tool.

## After Auditing

1. **Present a summary** to the user showing which docs are up to date and which have issues
2. **For each discrepancy**, explain what changed and propose an update
3. **Only apply updates with user approval** — don't silently rewrite docs
4. **Update `agent-docs/features/INDEX.md`** if new docs were added or existing ones renamed

## Creating a New Feature Doc

When creating a new feature doc:

1. **Research the feature** using Explore subagents to understand the behavior
2. **Write from the user's perspective** — describe what the feature does, not how it's coded
3. **Include permissions** that gate the feature (permission strings are part of the design)
4. **Note non-obvious interactions** with other features (e.g., shared messageBodyId across echoes)
5. **Add the new doc to `agent-docs/features/INDEX.md`**

## Verification Checklist

When auditing, verify these specific things:

- [ ] All permission strings mentioned exist in `permissions.go`
- [ ] Permission strings use hyphens (not underscores) per project convention
- [ ] Described user-facing behaviors match what the code actually does
- [ ] Stated interactions between features are accurate
- [ ] The INDEX.md links are correct and all docs are listed
