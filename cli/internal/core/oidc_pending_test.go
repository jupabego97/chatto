package core

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"hmans.de/chatto/internal/config"
)

func TestPendingOIDCIdentityLifecycle(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	token, err := core.CreatePendingOIDCIdentity(ctx, PendingOIDCIdentity{
		ProviderID:    "hub",
		ProviderLabel: "Chatto Hub",
		Issuer:        "https://issuer.example",
		Subject:       "subject-1",
		Email:         "USER@EXAMPLE.COM",
		EmailVerified: true,
		Name:          "Example User",
		Username:      "example-user",
		RedirectURL:   "/chat",
		Mode:          "login",
		CreatedAt:     time.Now().Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("CreatePendingOIDCIdentity: %v", err)
	}

	pending, err := core.GetPendingOIDCIdentity(ctx, token)
	if err != nil {
		t.Fatalf("GetPendingOIDCIdentity: %v", err)
	}
	if pending.ProviderID != "hub" || pending.Email != "user@example.com" || pending.CreatedAt.IsZero() {
		t.Fatalf("pending OIDC identity = %+v", pending)
	}

	consumed, err := core.ConsumePendingOIDCIdentity(ctx, token)
	if err != nil {
		t.Fatalf("ConsumePendingOIDCIdentity: %v", err)
	}
	if consumed.Subject != "subject-1" {
		t.Fatalf("consumed subject = %q, want subject-1", consumed.Subject)
	}
	if _, err := core.GetPendingOIDCIdentity(ctx, token); !errors.Is(err, ErrPendingOIDCNotFound) {
		t.Fatalf("GetPendingOIDCIdentity after consume error = %v, want ErrPendingOIDCNotFound", err)
	}
}

func TestPendingOIDCIdentityExpired(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	token, err := core.CreatePendingOIDCIdentity(ctx, PendingOIDCIdentity{
		ProviderID: "hub",
		Issuer:     "https://issuer.example",
		Subject:    "subject-expired",
	})
	if err != nil {
		t.Fatalf("CreatePendingOIDCIdentity: %v", err)
	}
	expired := PendingOIDCIdentity{
		ProviderID: "hub",
		Issuer:     "https://issuer.example",
		Subject:    "subject-expired",
		CreatedAt:  time.Now().Add(-PendingOIDCTokenTTL - time.Minute),
	}
	data, err := json.Marshal(expired)
	if err != nil {
		t.Fatalf("marshal expired pending identity: %v", err)
	}
	if _, err := core.storage.runtimeStateKV.Put(ctx, core.pendingOIDCTokenKey(token), data); err != nil {
		t.Fatalf("overwrite pending identity: %v", err)
	}

	if _, err := core.GetPendingOIDCIdentity(ctx, token); !errors.Is(err, ErrPendingOIDCExpired) {
		t.Fatalf("GetPendingOIDCIdentity error = %v, want ErrPendingOIDCExpired", err)
	}
	if _, err := core.GetPendingOIDCIdentity(ctx, token); !errors.Is(err, ErrPendingOIDCNotFound) {
		t.Fatalf("GetPendingOIDCIdentity after expiry cleanup error = %v, want ErrPendingOIDCNotFound", err)
	}
}

func TestProvisionOIDCUserCreatesPasswordlessUserWithoutEmail(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	user, created, err := core.ProvisionOIDCUser(ctx, OIDCProvisionProfile{
		ProviderID:    "hub",
		ProviderLabel: "Chatto Hub",
		Issuer:        "https://issuer.example",
		Subject:       "subject-passwordless",
		Name:          "OIDC User",
		Username:      "oidc_user",
	})
	if err != nil {
		t.Fatalf("ProvisionOIDCUser: %v", err)
	}
	if !created {
		t.Fatal("ProvisionOIDCUser created = false, want true")
	}
	if user.Login != "oidc_user" || user.DisplayName != "OIDC User" {
		t.Fatalf("provisioned user = %+v", user)
	}
	if len(core.Users.VerifiedEmails(user.Id)) != 0 {
		t.Fatalf("verified emails = %v, want none", core.Users.VerifiedEmails(user.Id))
	}

	found, err := core.GetUserByExternalIdentity(ctx, "https://issuer.example", "subject-passwordless")
	if err != nil {
		t.Fatalf("GetUserByExternalIdentity: %v", err)
	}
	if found == nil || found.Id != user.Id {
		t.Fatalf("external identity user = %v, want %s", found, user.Id)
	}
}

func TestProvisionOIDCUserReturnsExistingLinkedUser(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	first, created, err := core.ProvisionOIDCUser(ctx, OIDCProvisionProfile{
		ProviderID: "hub",
		Issuer:     "https://issuer.example",
		Subject:    "subject-existing",
		Username:   "existing-oidc",
	})
	if err != nil {
		t.Fatalf("first ProvisionOIDCUser: %v", err)
	}
	if !created {
		t.Fatal("first ProvisionOIDCUser created = false, want true")
	}

	second, created, err := core.ProvisionOIDCUser(ctx, OIDCProvisionProfile{
		ProviderID: "hub",
		Issuer:     "https://issuer.example",
		Subject:    "subject-existing",
		Username:   "different-claim",
	})
	if err != nil {
		t.Fatalf("second ProvisionOIDCUser: %v", err)
	}
	if created {
		t.Fatal("second ProvisionOIDCUser created = true, want false")
	}
	if second == nil || second.Id != first.Id {
		t.Fatalf("second user = %v, want %s", second, first.Id)
	}
}

func TestProvisionOIDCUserAttachesVerifiedEmailAndPromotesOwner(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)
	core.config.Owners = config.OwnersConfig{Emails: []string{"owner@example.com"}}

	user, created, err := core.ProvisionOIDCUser(ctx, OIDCProvisionProfile{
		ProviderID:    "hub",
		Issuer:        "https://issuer.example",
		Subject:       "subject-owner",
		Email:         "Owner@Example.com",
		EmailVerified: true,
		Name:          "OIDC Owner",
		Username:      "oidc-owner",
		ProviderLabel: "Chatto Hub",
	})
	if err != nil {
		t.Fatalf("ProvisionOIDCUser: %v", err)
	}
	if !created {
		t.Fatal("ProvisionOIDCUser created = false, want true")
	}

	found, err := core.GetUserByVerifiedEmail(ctx, "owner@example.com")
	if err != nil {
		t.Fatalf("GetUserByVerifiedEmail: %v", err)
	}
	if found.Id != user.Id {
		t.Fatalf("verified email user = %s, want %s", found.Id, user.Id)
	}
	isOwner, err := core.IsServerOwner(ctx, user.Id)
	if err != nil {
		t.Fatalf("IsServerOwner: %v", err)
	}
	if !isOwner {
		t.Fatal("OIDC verified owner email did not promote user to owner")
	}
}

func TestProvisionOIDCUserRollsBackWhenVerifiedEmailIsClaimed(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	existing, err := core.CreateUser(ctx, SystemActorID, "claimed-email-user", "Claimed", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, existing.Id, "claimed@example.com"); err != nil {
		t.Fatalf("AddVerifiedEmailDirect: %v", err)
	}

	_, created, err := core.ProvisionOIDCUser(ctx, OIDCProvisionProfile{
		ProviderID:    "hub",
		Issuer:        "https://issuer.example",
		Subject:       "subject-email-conflict",
		Email:         "claimed@example.com",
		EmailVerified: true,
		Username:      "claimed-email-oidc",
		ProviderLabel: "Chatto Hub",
	})
	if !errors.Is(err, ErrEmailAlreadyVerified) {
		t.Fatalf("ProvisionOIDCUser error = %v, want ErrEmailAlreadyVerified", err)
	}
	if created {
		t.Fatal("ProvisionOIDCUser created = true, want false on email conflict")
	}
	if _, err := core.GetUserByLogin(ctx, "claimed-email-oidc"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetUserByLogin rollback error = %v, want ErrNotFound", err)
	}
	found, err := core.GetUserByExternalIdentity(ctx, "https://issuer.example", "subject-email-conflict")
	if err != nil {
		t.Fatalf("GetUserByExternalIdentity: %v", err)
	}
	if found != nil {
		t.Fatalf("external identity after rollback = %v, want nil", found)
	}
}

func TestLinkOIDCIdentityToUser(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	user, err := core.CreateUser(ctx, SystemActorID, "local-user", "Local User", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	linked, err := core.LinkOIDCIdentityToUser(ctx, user.Id, OIDCProvisionProfile{
		ProviderID: "hub",
		Issuer:     "https://issuer.example",
		Subject:    "subject-link",
		Email:      "local@example.com",
		Username:   "local-user",
	})
	if err != nil {
		t.Fatalf("LinkOIDCIdentityToUser: %v", err)
	}
	if linked.Id != user.Id {
		t.Fatalf("linked user = %s, want %s", linked.Id, user.Id)
	}
	if len(core.Users.VerifiedEmails(user.Id)) != 0 {
		t.Fatalf("unverified provider email should not be attached: %v", core.Users.VerifiedEmails(user.Id))
	}

	other, err := core.CreateUser(ctx, SystemActorID, "other-local-user", "Other Local User", "password123")
	if err != nil {
		t.Fatalf("CreateUser other: %v", err)
	}
	_, err = core.LinkOIDCIdentityToUser(ctx, other.Id, OIDCProvisionProfile{
		ProviderID: "hub",
		Issuer:     "https://issuer.example",
		Subject:    "subject-link",
	})
	if !errors.Is(err, ErrExternalIdentityAlreadyClaimed) {
		t.Fatalf("LinkOIDCIdentityToUser conflict error = %v, want ErrExternalIdentityAlreadyClaimed", err)
	}
}

func TestLinkOIDCIdentityToUserAttachesUnclaimedVerifiedEmail(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	user, err := core.CreateUser(ctx, SystemActorID, "verified-email-link-user", "Verified Email Link User", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	linked, err := core.LinkOIDCIdentityToUser(ctx, user.Id, OIDCProvisionProfile{
		ProviderID:    "hub",
		Issuer:        "https://issuer.example",
		Subject:       "subject-verified-email-link",
		Email:         "Verified-Link@Example.com",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("LinkOIDCIdentityToUser: %v", err)
	}
	if linked.Id != user.Id {
		t.Fatalf("linked user = %s, want %s", linked.Id, user.Id)
	}
	found, err := core.GetUserByVerifiedEmail(ctx, "verified-link@example.com")
	if err != nil {
		t.Fatalf("GetUserByVerifiedEmail: %v", err)
	}
	if found.Id != user.Id {
		t.Fatalf("verified email user = %s, want %s", found.Id, user.Id)
	}
}

func TestLinkOIDCIdentityToUserDoesNotLinkWhenVerifiedEmailIsClaimed(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	claimed, err := core.CreateUser(ctx, SystemActorID, "claimed-local-user", "Claimed", "password123")
	if err != nil {
		t.Fatalf("CreateUser claimed: %v", err)
	}
	if err := core.AddVerifiedEmailDirect(ctx, claimed.Id, "claimed@example.com"); err != nil {
		t.Fatalf("AddVerifiedEmailDirect: %v", err)
	}
	target, err := core.CreateUser(ctx, SystemActorID, "target-local-user", "Target", "password123")
	if err != nil {
		t.Fatalf("CreateUser target: %v", err)
	}

	_, err = core.LinkOIDCIdentityToUser(ctx, target.Id, OIDCProvisionProfile{
		ProviderID:    "hub",
		Issuer:        "https://issuer.example",
		Subject:       "subject-email-link-conflict",
		Email:         "claimed@example.com",
		EmailVerified: true,
	})
	if !errors.Is(err, ErrEmailAlreadyVerified) {
		t.Fatalf("LinkOIDCIdentityToUser error = %v, want ErrEmailAlreadyVerified", err)
	}
	found, err := core.GetUserByExternalIdentity(ctx, "https://issuer.example", "subject-email-link-conflict")
	if err != nil {
		t.Fatalf("GetUserByExternalIdentity: %v", err)
	}
	if found != nil {
		t.Fatalf("external identity after failed link = %v, want nil", found)
	}
}

func TestOIDCLoginFallbackIsStableAndValid(t *testing.T) {
	profile := OIDCProvisionProfile{
		ProviderID: strings.Repeat("provider", 20),
		Issuer:     "https://issuer.example",
		Subject:    "opaque subject with spaces",
	}

	login := oidcStableFallbackLogin(profile)
	if len(login) > MaxLoginLength {
		t.Fatalf("fallback login len = %d, want <= %d", len(login), MaxLoginLength)
	}
	if err := ValidateLogin(login); err != nil {
		t.Fatalf("fallback login %q is invalid: %v", login, err)
	}
}
