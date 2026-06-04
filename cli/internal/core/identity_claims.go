package core

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"hmans.de/chatto/internal/config"
	"hmans.de/chatto/internal/events"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

var (
	ErrIdentityClaimTaken         = errors.New("identity claim is already owned by another user")
	ErrIdentityClaimOwnerMismatch = errors.New("identity claim is owned by another user")
	ErrIdentityClaimKeyring       = errors.New("identity claim keyring is not configured")
)

func (c *ChattoCore) loginIdentityClaimID(login string) (string, error) {
	return c.activeIdentityClaimID("login", strings.ToLower(strings.TrimSpace(login)))
}

func (c *ChattoCore) loginIdentityClaimIDs(login string) ([]string, error) {
	return c.identityClaimIDs("login", strings.ToLower(strings.TrimSpace(login)))
}

func (c *ChattoCore) emailIdentityClaimID(email string) (string, error) {
	return c.activeIdentityClaimID("email", strings.ToLower(strings.TrimSpace(email)))
}

func (c *ChattoCore) emailIdentityClaimIDs(email string) ([]string, error) {
	return c.identityClaimIDs("email", strings.ToLower(strings.TrimSpace(email)))
}

func (c *ChattoCore) oidcIdentityClaimID(issuer, subject string) (string, error) {
	return c.activeIdentityClaimID("oidc", issuer+"\x00"+subject)
}

func (c *ChattoCore) oidcIdentityClaimIDs(issuer, subject string) ([]string, error) {
	return c.identityClaimIDs("oidc", issuer+"\x00"+subject)
}

func (c *ChattoCore) activeIdentityClaimID(prefix, normalizedValue string) (string, error) {
	key, err := c.activeIdentityClaimKey()
	if err != nil {
		return "", err
	}
	return identityClaimID(prefix, normalizedValue, key)
}

func (c *ChattoCore) identityClaimIDs(prefix, normalizedValue string) ([]string, error) {
	keys, err := c.identityClaimKeysActiveFirst()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(keys))
	for _, key := range keys {
		id, err := identityClaimID(prefix, normalizedValue, key)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *ChattoCore) activeIdentityClaimKey() (config.IdentityClaimKeyConfig, error) {
	if c.config.IdentityClaims.ActiveKeyID == "" {
		return config.IdentityClaimKeyConfig{}, ErrIdentityClaimKeyring
	}
	for _, key := range c.config.IdentityClaims.Keys {
		if key.ID == c.config.IdentityClaims.ActiveKeyID {
			if key.Secret == "" {
				return config.IdentityClaimKeyConfig{}, ErrIdentityClaimKeyring
			}
			return key, nil
		}
	}
	return config.IdentityClaimKeyConfig{}, ErrIdentityClaimKeyring
}

func (c *ChattoCore) identityClaimKeysActiveFirst() ([]config.IdentityClaimKeyConfig, error) {
	active, err := c.activeIdentityClaimKey()
	if err != nil {
		return nil, err
	}
	keys := []config.IdentityClaimKeyConfig{active}
	for _, key := range c.config.IdentityClaims.Keys {
		if key.ID == active.ID {
			continue
		}
		if key.ID == "" || key.Secret == "" {
			return nil, ErrIdentityClaimKeyring
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func identityClaimID(prefix, normalizedValue string, key config.IdentityClaimKeyConfig) (string, error) {
	if prefix == "" || normalizedValue == "" || key.ID == "" || key.Secret == "" {
		return "", ErrIdentityClaimKeyring
	}
	mac := hmac.New(sha256.New, []byte(key.Secret))
	_, _ = mac.Write([]byte("chatto.identity_claim.v1"))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(prefix))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(normalizedValue))
	return prefix + "_" + key.ID + "_" + hex.EncodeToString(mac.Sum(nil)), nil
}

func (c *ChattoCore) claimIdentity(ctx context.Context, actorID, ownerUserID, claimID string, kind corev1.IdentityClaimKind) (uint64, error) {
	for attempt := 0; attempt < maxUserMutationRetries; attempt++ {
		state, err := c.readIdentityClaimState(ctx, claimID)
		if err != nil {
			return 0, err
		}
		if state.active {
			if state.ownerUserID == ownerUserID && state.kind == kind {
				return state.lastSeq, nil
			}
			return 0, ErrIdentityClaimTaken
		}

		event := newEvent(actorID, &corev1.Event{Event: &corev1.Event_IdentityClaimed{
			IdentityClaimed: &corev1.IdentityClaimedEvent{
				ClaimId:     claimID,
				OwnerUserId: ownerUserID,
				Kind:        kind,
			},
		}})
		agg := events.IdentityClaimAggregate(claimID)
		seq, err := c.EventPublisher.AppendAtFilter(ctx, agg.SubjectFor(event), event, agg.AllEventsFilter(), state.lastSeq)
		if err == nil {
			return seq, nil
		}
		if !errors.Is(err, events.ErrConflict) {
			return 0, err
		}
		if err := waitIdentityClaimRetry(ctx, attempt); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("identity claim OCC retry exhausted after %d attempts: %w", maxUserMutationRetries, events.ErrConflict)
}

func (c *ChattoCore) releaseIdentity(ctx context.Context, actorID, ownerUserID, claimID string, kind corev1.IdentityClaimKind) (uint64, error) {
	for attempt := 0; attempt < maxUserMutationRetries; attempt++ {
		state, err := c.readIdentityClaimState(ctx, claimID)
		if err != nil {
			return 0, err
		}
		if !state.active {
			return state.lastSeq, nil
		}
		if state.ownerUserID != ownerUserID || state.kind != kind {
			return 0, ErrIdentityClaimOwnerMismatch
		}

		event := newEvent(actorID, &corev1.Event{Event: &corev1.Event_IdentityClaimReleased{
			IdentityClaimReleased: &corev1.IdentityClaimReleasedEvent{
				ClaimId:     claimID,
				OwnerUserId: ownerUserID,
				Kind:        kind,
			},
		}})
		agg := events.IdentityClaimAggregate(claimID)
		seq, err := c.EventPublisher.AppendAtFilter(ctx, agg.SubjectFor(event), event, agg.AllEventsFilter(), state.lastSeq)
		if err == nil {
			return seq, nil
		}
		if !errors.Is(err, events.ErrConflict) {
			return 0, err
		}
		if err := waitIdentityClaimRetry(ctx, attempt); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("identity release OCC retry exhausted after %d attempts: %w", maxUserMutationRetries, events.ErrConflict)
}

func waitIdentityClaimRetry(ctx context.Context, attempt int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(1<<attempt) * time.Millisecond):
		return nil
	}
}

type identityClaimState struct {
	active      bool
	ownerUserID string
	kind        corev1.IdentityClaimKind
	lastSeq     uint64
}

func (c *ChattoCore) readIdentityClaimState(ctx context.Context, claimID string) (identityClaimState, error) {
	agg := events.IdentityClaimAggregate(claimID)
	published, lastSeq, err := c.EventPublisher.SubjectEvents(ctx, agg.AllEventsFilter())
	if err != nil {
		return identityClaimState{}, fmt.Errorf("read identity claim events: %w", err)
	}

	state := identityClaimState{lastSeq: lastSeq}
	for _, event := range published {
		switch e := event.GetEvent().(type) {
		case *corev1.Event_IdentityClaimed:
			claim := e.IdentityClaimed
			if claim.GetClaimId() != claimID {
				continue
			}
			state.active = true
			state.ownerUserID = claim.GetOwnerUserId()
			state.kind = claim.GetKind()
		case *corev1.Event_IdentityClaimReleased:
			release := e.IdentityClaimReleased
			if release.GetClaimId() != claimID {
				continue
			}
			state.active = false
			state.ownerUserID = release.GetOwnerUserId()
			state.kind = release.GetKind()
		}
	}
	return state, nil
}
