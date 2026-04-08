package core

import (
	"errors"
	"testing"

	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

func TestEventPublishingHelpers_RejectInvalidEvents(t *testing.T) {
	core := &ChattoCore{}
	ctx := testContext(t)

	t.Run("publishSpaceEvent rejects nil pointer", func(t *testing.T) {
		err := core.publishSpaceEvent(ctx, "space.test", nil)
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected ErrInvalidEvent, got: %v", err)
		}
	})

	t.Run("publishSpaceEvent rejects unset oneof payload", func(t *testing.T) {
		err := core.publishSpaceEvent(ctx, "space.test", &corev1.SpaceEvent{})
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected ErrInvalidEvent, got: %v", err)
		}
	})

	t.Run("publishLiveSpaceEvent rejects invalid payload", func(t *testing.T) {
		err := core.publishLiveSpaceEvent(ctx, "live.space.test", &corev1.SpaceEvent{})
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected ErrInvalidEvent, got: %v", err)
		}
	})

	t.Run("publishInstanceEvent rejects invalid payload", func(t *testing.T) {
		err := core.publishInstanceEvent(ctx, "live.instance.test", &corev1.InstanceEvent{})
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected ErrInvalidEvent, got: %v", err)
		}
	})

	t.Run("publishSpaceEventWithAck rejects invalid payload", func(t *testing.T) {
		seq, err := core.publishSpaceEventWithAck(ctx, "space.test", &corev1.SpaceEvent{})
		if seq != 0 {
			t.Fatalf("expected sequence 0 on error, got: %d", seq)
		}
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected ErrInvalidEvent, got: %v", err)
		}
	})

	t.Run("publishSpaceEventWithOCC rejects invalid payload", func(t *testing.T) {
		seq, err := core.publishSpaceEventWithOCC(ctx, "space123", "space.test", &corev1.SpaceEvent{})
		if seq != 0 {
			t.Fatalf("expected sequence 0 on error, got: %d", seq)
		}
		if !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected ErrInvalidEvent, got: %v", err)
		}
	})
}
