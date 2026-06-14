package core

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

func suspensionEvent(eventID, actorID, userID, reason string, createdAt time.Time, expiresAt *time.Time) *corev1.Event {
	payload := &corev1.UserSuspendedEvent{
		UserId: userID,
		Reason: reason,
	}
	if expiresAt != nil {
		payload.ExpiresAt = timestamppb.New(*expiresAt)
	}
	return userEvent(eventID, createdAt, &corev1.Event{
		ActorId: actorID,
		Event:   &corev1.Event_UserSuspended{UserSuspended: payload},
	})
}

func unsuspensionEvent(eventID, actorID, userID, reason string, createdAt time.Time) *corev1.Event {
	return userEvent(eventID, createdAt, &corev1.Event{
		ActorId: actorID,
		Event: &corev1.Event_UserUnsuspended{UserUnsuspended: &corev1.UserUnsuspendedEvent{
			UserId: userID,
			Reason: reason,
		}},
	})
}

func TestUserSuspensionProjection(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	t.Run("active suspension", func(t *testing.T) {
		p := NewUserSuspensionProjection()
		expiresAt := now.Add(time.Hour)
		if err := p.Apply(suspensionEvent("E1", "Umod", "Utarget", "cool down", now, &expiresAt), 1); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		suspension, ok := p.ActiveSuspension("Utarget", now)
		if !ok {
			t.Fatal("expected active suspension")
		}
		if suspension.EventID != "E1" || suspension.ModeratorID != "Umod" || suspension.Reason != "cool down" {
			t.Fatalf("unexpected suspension: %#v", suspension)
		}
	})

	t.Run("expired suspension stops applying", func(t *testing.T) {
		p := NewUserSuspensionProjection()
		expiresAt := now.Add(-time.Minute)
		if err := p.Apply(suspensionEvent("E1", "Umod", "Utarget", "expired", now.Add(-time.Hour), &expiresAt), 1); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		if p.IsActive("Utarget", now) {
			t.Fatal("expected expired suspension to be inactive")
		}
	})

	t.Run("later suspension supersedes earlier suspension", func(t *testing.T) {
		p := NewUserSuspensionProjection()
		firstExpiry := now.Add(time.Hour)
		secondExpiry := now.Add(2 * time.Hour)
		if err := p.Apply(suspensionEvent("E1", "Umod", "Utarget", "first", now, &firstExpiry), 1); err != nil {
			t.Fatalf("Apply first: %v", err)
		}
		if err := p.Apply(suspensionEvent("E2", "Umod2", "Utarget", "second", now.Add(time.Minute), &secondExpiry), 2); err != nil {
			t.Fatalf("Apply second: %v", err)
		}
		suspension, ok := p.ActiveSuspension("Utarget", now)
		if !ok {
			t.Fatal("expected active suspension")
		}
		if suspension.EventID != "E2" || suspension.ModeratorID != "Umod2" || suspension.Reason != "second" {
			t.Fatalf("expected later suspension, got %#v", suspension)
		}
	})

	t.Run("unsuspend clears active suspension", func(t *testing.T) {
		p := NewUserSuspensionProjection()
		if err := p.Apply(suspensionEvent("E1", "Umod", "Utarget", "active", now, nil), 1); err != nil {
			t.Fatalf("Apply suspension: %v", err)
		}
		if err := p.Apply(unsuspensionEvent("E2", "Umod", "Utarget", "served", now.Add(time.Minute)), 2); err != nil {
			t.Fatalf("Apply unsuspension: %v", err)
		}
		if p.IsActive("Utarget", now.Add(2*time.Minute)) {
			t.Fatal("expected unsuspended user to be inactive")
		}
	})

	t.Run("deleted user clears active suspension", func(t *testing.T) {
		p := NewUserSuspensionProjection()
		if err := p.Apply(suspensionEvent("E1", "Umod", "Utarget", "active", now, nil), 1); err != nil {
			t.Fatalf("Apply suspension: %v", err)
		}
		if err := p.Apply(userEvent("E2", now.Add(time.Minute), &corev1.Event{
			Event: &corev1.Event_UserAccountDeleted{UserAccountDeleted: &corev1.UserAccountDeletedEvent{
				UserId: "Utarget",
			}},
		}), 2); err != nil {
			t.Fatalf("Apply delete: %v", err)
		}
		if p.IsActive("Utarget", now.Add(2*time.Minute)) {
			t.Fatal("expected deleted user suspension to be cleared")
		}
	})
}
