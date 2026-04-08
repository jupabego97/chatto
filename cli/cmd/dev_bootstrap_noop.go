//go:build !dev

package cmd

import (
	"context"

	"hmans.de/chatto/internal/core"
)

func init() {
	// No-op in production builds
	devStartupHook = func(context.Context, *core.ChattoCore) {}
}
