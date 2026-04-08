---
# chatto-lhwo
title: Extract RBAC engine to pkg/rbac
status: draft
type: task
priority: low
created_at: 2026-02-28T12:30:38Z
updated_at: 2026-02-28T12:30:38Z
parent: chatto-w2dd
---

Extract the generic RBAC engine from cli/internal/core/rbac into cli/pkg/rbac.

## Scope

The engine is explicitly designed to be generic (used for both instance-level and space-level RBAC). Supports:
- Role CRUD with hierarchy (position-based ordering)
- Permission grant/deny per role
- Role assignments to users
- Hierarchy-wins resolution

## Blockers

The engine currently serializes roles using Chatto's protobuf-generated types (corev1.Role, etc.). To extract cleanly, we'd need to either:
1. Define local Role/Permission structs in the pkg and convert at the boundary
2. Use a generic serialization interface ([]byte + marshal/unmarshal funcs)
3. Accept the proto dependency (makes the pkg less portable)

This needs design discussion before implementation.

## Tasks
- [ ] Design the decoupling approach for proto types
- [ ] Create cli/pkg/rbac package
- [ ] Migrate the engine
- [ ] Update callers in core
- [ ] Verify tests pass
