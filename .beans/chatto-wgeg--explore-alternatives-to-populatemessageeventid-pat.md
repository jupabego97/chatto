---
# chatto-wgeg
title: Explore alternatives to populateMessageEventId pattern
status: draft
type: task
created_at: 2026-01-18T10:43:31Z
updated_at: 2026-01-18T10:43:31Z
---

The current pattern copies Event.Id to inner MessagePostedEvent.EventId at read time so field resolvers can access it. This feels awkward because the Event wrapper already contains the ID.

Potential alternatives:
1. Store parent Event in context before resolving child fields - adds hidden coupling but removes duplication
2. Change GraphQL schema to not unwrap the union - breaks current API
3. Keep current approach - pragmatic, explicit, minimal overhead

Related to the reactions refactoring (chatto-ptj6) where this pattern was introduced for messageEventId lookups.