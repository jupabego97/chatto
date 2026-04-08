package model

// Viewer represents the current authenticated user's instance-level permissions.
// UserID and IsConfigAdmin are internal fields used by field resolvers.
type Viewer struct {
	UserID        string
	IsConfigAdmin bool
}
