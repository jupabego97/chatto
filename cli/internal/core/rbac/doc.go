// Package rbac provides a generic RBAC (Role-Based Access Control) engine
// that can be used by both instance-level and space-level authorization systems.
//
// The engine provides:
//   - Role CRUD operations (create, get, list, update, delete)
//   - Permission grants/revokes per role
//   - Role assignments per user
//   - Optional user-level permission overrides (grant/deny)
//
// Each system (instance or space) wraps the engine with an adapter that handles
// scope-specific logic like implicit roles or storage bucket selection.
package rbac
