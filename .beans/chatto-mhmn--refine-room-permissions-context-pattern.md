---
# chatto-mhmn
title: Refine room permissions context pattern
status: draft
type: task
created_at: 2026-01-15T16:53:01Z
updated_at: 2026-01-15T16:53:01Z
---

Follow-up from the initial roomPermissions context implementation. Consider:

- Moving canPostMessage to context as well (currently still prop-drilled to ChatInput)
- Evaluating if we need a more general 'can' API pattern similar to backend's can.go
- Consider if other permissions should be added to the context
- Review if the reactive state wrapper pattern is the best approach or if there's a cleaner Svelte 5 idiom

The current implementation works but was kept minimal. This task is for future refinement when we have more context usage patterns to inform the design.