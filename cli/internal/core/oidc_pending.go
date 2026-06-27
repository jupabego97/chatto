package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"hmans.de/chatto/internal/events"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

const (
	PendingOIDCTokenTTL       = 15 * time.Minute
	pendingOIDCTokenKeyPrefix = "pending_oidc."
)

var (
	ErrPendingOIDCNotFound = errors.New("pending OIDC login not found")
	ErrPendingOIDCExpired  = errors.New("pending OIDC login has expired")
)

type PendingOIDCIdentity struct {
	ProviderID    string    `json:"provider_id"`
	ProviderLabel string    `json:"provider_label"`
	Issuer        string    `json:"issuer"`
	Subject       string    `json:"subject"`
	Email         string    `json:"email,omitempty"`
	EmailVerified bool      `json:"email_verified,omitempty"`
	Name          string    `json:"name,omitempty"`
	Username      string    `json:"username,omitempty"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	RedirectURL   string    `json:"redirect_url,omitempty"`
	Mode          string    `json:"mode,omitempty"`
	LinkUserID    string    `json:"link_user_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type OIDCProvisionProfile struct {
	ProviderID    string
	ProviderLabel string
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Username      string
}

func (c *ChattoCore) pendingOIDCTokenKey(token string) string {
	return c.runtimeTokenKey(pendingOIDCTokenKeyPrefix, token)
}

func (c *ChattoCore) CreatePendingOIDCIdentity(ctx context.Context, pending PendingOIDCIdentity) (string, error) {
	if pending.ProviderID == "" || pending.Issuer == "" || pending.Subject == "" {
		return "", fmt.Errorf("provider ID, issuer, and subject are required")
	}
	token := NewPendingOIDCToken()
	pending.CreatedAt = time.Now()
	if pending.EmailVerified {
		pending.Email = strings.ToLower(strings.TrimSpace(pending.Email))
	} else {
		pending.Email = ""
	}
	data, err := json.Marshal(pending)
	if err != nil {
		return "", fmt.Errorf("marshal pending OIDC identity: %w", err)
	}
	if _, err := c.storage.runtimeStateKV.Create(ctx, c.pendingOIDCTokenKey(token), data, jetstream.KeyTTL(PendingOIDCTokenTTL)); err != nil {
		return "", fmt.Errorf("store pending OIDC identity: %w", err)
	}
	return token, nil
}

func (c *ChattoCore) GetPendingOIDCIdentity(ctx context.Context, token string) (*PendingOIDCIdentity, error) {
	key := c.pendingOIDCTokenKey(token)
	entry, err := c.storage.runtimeStateKV.Get(ctx, key)
	if err != nil {
		if isRuntimeStateKeyAbsent(err) {
			return nil, ErrPendingOIDCNotFound
		}
		return nil, fmt.Errorf("get pending OIDC identity: %w", err)
	}
	var pending PendingOIDCIdentity
	if err := json.Unmarshal(entry.Value(), &pending); err != nil {
		return nil, fmt.Errorf("unmarshal pending OIDC identity: %w", err)
	}
	if time.Since(pending.CreatedAt) > PendingOIDCTokenTTL {
		_ = c.storage.runtimeStateKV.Delete(ctx, key)
		return nil, ErrPendingOIDCExpired
	}
	return &pending, nil
}

func (c *ChattoCore) ConsumePendingOIDCIdentity(ctx context.Context, token string) (*PendingOIDCIdentity, error) {
	pending, err := c.GetPendingOIDCIdentity(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := c.DeletePendingOIDCIdentity(ctx, token); err != nil {
		return nil, err
	}
	return pending, nil
}

func (c *ChattoCore) DeletePendingOIDCIdentity(ctx context.Context, token string) error {
	err := c.storage.runtimeStateKV.Delete(ctx, c.pendingOIDCTokenKey(token))
	if err != nil && !isRuntimeStateKeyAbsent(err) {
		return fmt.Errorf("delete pending OIDC identity: %w", err)
	}
	return nil
}

func (c *ChattoCore) ProvisionOIDCUser(ctx context.Context, profile OIDCProvisionProfile) (*corev1.User, bool, error) {
	if existing, err := c.GetUserByExternalIdentity(ctx, profile.Issuer, profile.Subject); err != nil || existing != nil {
		return existing, false, err
	}

	login, err := c.availableOIDCLogin(ctx, profile)
	if err != nil {
		return nil, false, err
	}
	displayName := oidcDisplayName(profile)
	user, err := c.CreateUser(ctx, SystemActorID, login, displayName, "")
	if err != nil {
		return nil, false, err
	}
	created := true
	rollback := func() {
		if created {
			c.rollbackUserCreation(ctx, user)
		}
	}

	if err := c.LinkExternalIdentity(ctx, profile.ProviderID, "oidc", profile.Issuer, profile.Subject, user.Id); err != nil {
		rollback()
		if errors.Is(err, ErrExternalIdentityAlreadyClaimed) {
			return nil, false, err
		}
		return nil, false, fmt.Errorf("link OIDC identity: %w", err)
	}

	if profile.EmailVerified && strings.TrimSpace(profile.Email) != "" {
		if err := c.AddVerifiedEmailDirect(ctx, user.Id, profile.Email); err != nil {
			rollback()
			return nil, false, fmt.Errorf("attach verified OIDC email: %w", err)
		}
	}

	return user, true, nil
}

func (c *ChattoCore) LinkOIDCIdentityToUser(ctx context.Context, userID string, profile OIDCProvisionProfile) (*corev1.User, error) {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	email := ""
	if profile.EmailVerified {
		email = strings.ToLower(strings.TrimSpace(profile.Email))
	}

	agg := events.UserAggregate(userID)
	entries := make([]events.BatchEntry, 0, 2)
	if email != "" {
		if existing, err := c.GetUserByVerifiedEmail(ctx, email); err != nil {
			if !errors.Is(err, ErrNotFound) {
				return nil, err
			}
			emailEvent := newEvent(userID, &corev1.Event{Event: &corev1.Event_UserVerifiedEmailAdded{
				UserVerifiedEmailAdded: &corev1.UserVerifiedEmailAddedEvent{UserId: userID},
			}})
			encryptedEmail, err := c.encryptUserPIIString(ctx, emailEvent.GetId(), userID, events.EventUserVerifiedEmailAdded, "email", email)
			if err != nil {
				return nil, fmt.Errorf("encrypt verified email: %w", err)
			}
			emailEvent.GetUserVerifiedEmailAdded().EncryptedEmail = encryptedEmail
			entries = append(entries, events.BatchEntry{
				Subject: agg.SubjectFor(emailEvent),
				Event:   emailEvent,
			})
		} else if existing.GetId() != userID {
			return nil, ErrEmailAlreadyVerified
		}
	}

	identityEvent := newEvent(userID, &corev1.Event{Event: &corev1.Event_UserExternalIdentityLinked{
		UserExternalIdentityLinked: &corev1.UserExternalIdentityLinkedEvent{
			UserId:       userID,
			Issuer:       profile.Issuer,
			Subject:      profile.Subject,
			SubjectHash:  externalIdentityHash(profile.Issuer, profile.Subject),
			ProviderId:   profile.ProviderID,
			ProviderType: "oidc",
		},
	}})
	entries = append(entries, events.BatchEntry{
		Subject: agg.SubjectFor(identityEvent),
		Event:   identityEvent,
	})

	if _, err := c.appendUserBatch(ctx, userID, entries, events.UserSubjectFilter(), func() error {
		if email != "" {
			if existing, ok := c.Users.GetByEmail(email); ok && existing.GetId() != userID {
				return ErrEmailAlreadyVerified
			}
		}
		if existing, ok := c.Users.GetByExternalIdentity(profile.Issuer, profile.Subject); ok && existing.GetId() != userID {
			return ErrExternalIdentityAlreadyClaimed
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if email != "" {
		c.assignOwnerRoleForVerifiedEmail(ctx, userID, email)
	}
	return user, nil
}

var oidcLoginCleanPattern = regexp.MustCompile(`[^a-z0-9._-]+`)

func (c *ChattoCore) availableOIDCLogin(ctx context.Context, profile OIDCProvisionProfile) (string, error) {
	for _, candidate := range oidcLoginCandidates(profile) {
		if _, err := c.GetUserByLogin(ctx, candidate); err != nil {
			if errors.Is(err, ErrNotFound) {
				return candidate, nil
			}
			return "", err
		}
	}
	fallback := oidcStableFallbackLogin(profile)
	for i := 0; ; i++ {
		candidate := fallback
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", fallback, i+1)
		}
		if _, err := c.GetUserByLogin(ctx, candidate); err != nil {
			if errors.Is(err, ErrNotFound) {
				return candidate, nil
			}
			return "", err
		}
	}
}

func oidcLoginCandidates(profile OIDCProvisionProfile) []string {
	raw := []string{profile.Username}
	if profile.EmailVerified && profile.Email != "" {
		local, _, ok := strings.Cut(profile.Email, "@")
		if ok {
			raw = append(raw, local)
		}
	}
	raw = append(raw, profile.Name)

	var out []string
	seen := map[string]struct{}{}
	for _, value := range raw {
		candidate := sanitizeOIDCLogin(value)
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		out = append(out, candidate)
	}
	out = append(out, oidcStableFallbackLogin(profile))
	return out
}

func sanitizeOIDCLogin(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = oidcLoginCleanPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-_")
	if len(value) > MaxLoginLength {
		value = strings.Trim(value[:MaxLoginLength], ".-_")
	}
	if value == "" || ValidateLogin(value) != nil {
		return ""
	}
	return value
}

func oidcStableFallbackLogin(profile OIDCProvisionProfile) string {
	provider := sanitizeOIDCLogin(profile.ProviderID)
	if provider == "" {
		provider = "oidc"
	}
	hash := externalIdentityHash(profile.Issuer, profile.Subject)
	maxProviderLen := MaxLoginLength - 1 - 10
	if len(provider) > maxProviderLen {
		provider = provider[:maxProviderLen]
	}
	return provider + "-" + hash[:10]
}

func oidcDisplayName(profile OIDCProvisionProfile) string {
	for _, value := range []string{profile.Name, profile.Username} {
		value = NormalizeDisplayName(value)
		if value != "" && ValidateDisplayName(value) == nil {
			return truncateRunes(value, MaxDisplayNameLength)
		}
	}
	if profile.EmailVerified && profile.Email != "" {
		local, _, ok := strings.Cut(profile.Email, "@")
		if ok {
			local = NormalizeDisplayName(local)
			if local != "" && ValidateDisplayName(local) == nil {
				return truncateRunes(local, MaxDisplayNameLength)
			}
		}
	}
	return oidcStableFallbackLogin(profile)
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
