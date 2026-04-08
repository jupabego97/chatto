package dataloader

import (
	"context"
)

// contextKey is an unexported type for context keys to prevent collisions
type contextKey struct {
	name string
}

var loadersCtxKey = &contextKey{"dataloaders"}

// ForContext extracts the dataloaders from the GraphQL context.
// Returns nil if no loaders are available (e.g., in tests without middleware).
func ForContext(ctx context.Context) *Loaders {
	raw, _ := ctx.Value(loadersCtxKey).(*Loaders)
	return raw
}

// WithLoaders returns a new context with the loaders injected
func WithLoaders(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersCtxKey, loaders)
}
