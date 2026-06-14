package core

import (
	"fmt"
	"sort"
	"time"

	"hmans.de/chatto/internal/events"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

type UserSuspension struct {
	EventID     string
	UserID      string
	ModeratorID string
	Reason      string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

func (s UserSuspension) Active(now time.Time) bool {
	return s.ExpiresAt == nil || s.ExpiresAt.After(now)
}

type UserSuspensionProjection struct {
	events.MemoryProjection
	byUser map[string]UserSuspension
}

func NewUserSuspensionProjection() *UserSuspensionProjection {
	return &UserSuspensionProjection{byUser: make(map[string]UserSuspension)}
}

func (p *UserSuspensionProjection) Subjects() []string {
	return []string{events.UserSubjectFilter()}
}

func (p *UserSuspensionProjection) Apply(event *corev1.Event, _ uint64) error {
	if event == nil {
		return nil
	}
	p.Lock()
	defer p.Unlock()

	switch e := event.GetEvent().(type) {
	case *corev1.Event_UserSuspended:
		suspended := e.UserSuspended
		userID := suspended.GetUserId()
		reason := suspended.GetReason()
		if userID == "" || reason == "" {
			return fmt.Errorf("UserSuspended missing userID or reason")
		}
		createdAt := time.Time{}
		if ts := event.GetCreatedAt(); ts != nil {
			createdAt = ts.AsTime()
		}
		var expiresAt *time.Time
		if ts := suspended.GetExpiresAt(); ts != nil {
			t := ts.AsTime()
			expiresAt = &t
		}
		p.byUser[userID] = UserSuspension{
			EventID:     event.GetId(),
			UserID:      userID,
			ModeratorID: event.GetActorId(),
			Reason:      reason,
			CreatedAt:   createdAt,
			ExpiresAt:   expiresAt,
		}
	case *corev1.Event_UserUnsuspended:
		delete(p.byUser, e.UserUnsuspended.GetUserId())
	case *corev1.Event_UserAccountDeleted:
		delete(p.byUser, e.UserAccountDeleted.GetUserId())
	default:
	}
	return nil
}

func (p *UserSuspensionProjection) ActiveSuspension(userID string, now time.Time) (UserSuspension, bool) {
	p.RLock()
	defer p.RUnlock()
	suspension, ok := p.byUser[userID]
	if !ok || !suspension.Active(now) {
		return UserSuspension{}, false
	}
	return suspension, true
}

func (p *UserSuspensionProjection) IsActive(userID string, now time.Time) bool {
	_, ok := p.ActiveSuspension(userID, now)
	return ok
}

func (p *UserSuspensionProjection) ActiveSuspensions(now time.Time) []UserSuspension {
	p.RLock()
	defer p.RUnlock()
	out := []UserSuspension{}
	for _, suspension := range p.byUser {
		if suspension.Active(now) {
			out = append(out, suspension)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (p *UserSuspensionProjection) adminProjectionEstimate() (entries int64, bytes int64, metrics []ProjectionAdminMetric) {
	p.RLock()
	defer p.RUnlock()
	entries = int64(len(p.byUser))
	for _, suspension := range p.byUser {
		bytes += int64(len(suspension.EventID) + len(suspension.UserID) + len(suspension.ModeratorID) + len(suspension.Reason) + 64)
	}
	metrics = []ProjectionAdminMetric{{Name: "tracked users", Value: entries}}
	return entries, bytes, metrics
}
