---
# chatto-b4no
title: Update ARCHITECTURE.md and authorization.md for new RBAC model
status: todo
type: task
priority: normal
created_at: 2026-01-29T08:31:17Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzz
---

Update documentation to reflect the new Space RBAC implementation from the `hmans/feat/space-role-assignments` branch.

## ARCHITECTURE.md Updates

- [ ] Update permission names throughout (e.g., `rooms.browse` → `room.list`, `roles.manage` → `role.manage`)
- [ ] Add room-scoped permissions section (message.post, message.react, message.edit.*, message.delete.*)
- [ ] Update INSTANCE_RBAC key patterns to new format (`allow.{role}.{verb}.{objectType}.{objectId}`)
- [ ] Remove stale RBAC keys from SPACE_CONFIG section (now in SPACE_RBAC)
- [ ] Update Can* function table with new functions and correct file references (space_can.go, instance_can.go)
- [ ] Update Instance Permissions table with correct permission names (dm.view, dm.write, user.delete, etc.)

## authorization.md Updates

- [ ] Rewrite "Permission Constant Naming" section for new unified Permission type
- [ ] Update Built-in Permissions table with correct dot-notation names
- [ ] Update Mutations table with new permission requirements (message.post, message.react, etc.)
- [ ] Update file references (can.go → space_can.go/instance_can.go, permissions.go → permission.go/permission_ops.go/permission_resolver.go)
- [ ] Add section explaining deny-always-wins, instance-authority-first model

## Reference

See code review in PR for full details of what changed.
