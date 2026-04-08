package dataloader

import (
	"context"

	"github.com/vikstrous/dataloadgen"
	"hmans.de/chatto/internal/core"
)

// ReactionKey identifies a specific message's reactions within a space.
type ReactionKey struct {
	SpaceID string
	EventID string
}

// newReactionLoader creates a dataloader that batches reaction lookups.
// Multiple messages' reactions are fetched in a single ListKeysFiltered call per space.
func newReactionLoader(c *core.ChattoCore) *dataloadgen.Loader[ReactionKey, []core.ReactionSummary] {
	return dataloadgen.NewLoader(
		func(ctx context.Context, keys []ReactionKey) ([][]core.ReactionSummary, []error) {
			return batchGetReactions(ctx, c, keys)
		},
		dataloadgen.WithWait(defaultWait),
	)
}

// batchGetReactions fetches reactions for multiple messages, grouped by space.
// Returns results and errors in the same order as keys.
func batchGetReactions(ctx context.Context, c *core.ChattoCore, keys []ReactionKey) ([][]core.ReactionSummary, []error) {
	results := make([][]core.ReactionSummary, len(keys))
	errs := make([]error, len(keys))

	// Group keys by spaceID for batch fetching
	type indexedEventID struct {
		index   int
		eventID string
	}
	bySpace := make(map[string][]indexedEventID)
	for i, k := range keys {
		bySpace[k.SpaceID] = append(bySpace[k.SpaceID], indexedEventID{index: i, eventID: k.EventID})
	}

	for spaceID, entries := range bySpace {
		eventIDs := make([]string, len(entries))
		for i, e := range entries {
			eventIDs[i] = e.eventID
		}

		batch, err := c.GetReactionsBatch(ctx, spaceID, eventIDs)
		if err != nil {
			for _, e := range entries {
				errs[e.index] = err
			}
			continue
		}

		for _, e := range entries {
			summaries := batch[e.eventID]
			if summaries == nil {
				summaries = []core.ReactionSummary{}
			}
			results[e.index] = summaries
		}
	}

	return results, errs
}
