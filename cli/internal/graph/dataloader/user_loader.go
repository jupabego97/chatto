package dataloader

import (
	"context"
	"time"

	"github.com/vikstrous/dataloadgen"
	"hmans.de/chatto/internal/core"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// defaultWait is the time to wait before batching.
// 2ms allows most concurrent resolvers to queue their requests.
const defaultWait = 2 * time.Millisecond

// newUserLoader creates a dataloader for users.
func newUserLoader(c *core.ChattoCore) *dataloadgen.Loader[string, *corev1.User] {
	return dataloadgen.NewLoader(
		func(ctx context.Context, userIDs []string) ([]*corev1.User, []error) {
			return batchGetUsers(ctx, c, userIDs)
		},
		dataloadgen.WithWait(defaultWait),
	)
}

// batchGetUsers fetches multiple users in a single batch.
// Returns users and errors in the same order as userIDs.
func batchGetUsers(ctx context.Context, c *core.ChattoCore, userIDs []string) ([]*corev1.User, []error) {
	users, err := c.GetUsers(ctx, userIDs)
	if err != nil {
		// Return the same error for all IDs
		errors := make([]error, len(userIDs))
		for i := range errors {
			errors[i] = err
		}
		return nil, errors
	}

	// Build a map for O(1) lookup
	userMap := make(map[string]*corev1.User, len(users))
	for _, user := range users {
		if user != nil {
			userMap[user.Id] = user
		}
	}

	// Return users in the same order as userIDs
	result := make([]*corev1.User, len(userIDs))
	errors := make([]error, len(userIDs))
	for i, id := range userIDs {
		if user, ok := userMap[id]; ok {
			result[i] = user
		} else {
			errors[i] = core.ErrNotFound
		}
	}

	return result, errors
}
