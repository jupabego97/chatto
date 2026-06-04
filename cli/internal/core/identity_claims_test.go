package core

import (
	"errors"
	"strings"
	"testing"

	"hmans.de/chatto/internal/config"
	"hmans.de/chatto/internal/events"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

func TestIdentityClaimIDsUseKeyedNormalizedTokens(t *testing.T) {
	core, _ := setupTestCore(t)
	core.config.IdentityClaims = testIdentityClaimsConfig()

	loginClaimID, err := core.loginIdentityClaimID(" Alice.Smith ")
	if err != nil {
		t.Fatalf("loginIdentityClaimID: %v", err)
	}
	loginClaimIDLower, err := core.loginIdentityClaimID("alice.smith")
	if err != nil {
		t.Fatalf("loginIdentityClaimID lower: %v", err)
	}
	if loginClaimID != loginClaimIDLower {
		t.Fatalf("login claim ID should be case-insensitive and trim whitespace")
	}
	if strings.Contains(loginClaimID, "alice") || strings.Contains(loginClaimID, ".") {
		t.Fatalf("login claim ID %q should not expose the login or add subject tokens", loginClaimID)
	}
	if !strings.HasPrefix(loginClaimID, "login_v2_") {
		t.Fatalf("login claim ID %q should include the active key ID", loginClaimID)
	}

	emailClaimID, err := core.emailIdentityClaimID("alice@example.com")
	if err != nil {
		t.Fatalf("emailIdentityClaimID: %v", err)
	}
	if strings.Contains(emailClaimID, "alice") || strings.Contains(emailClaimID, "@") {
		t.Fatalf("email claim ID %q should not expose the email address", emailClaimID)
	}

	oidcClaimID, err := core.oidcIdentityClaimID("https://issuer.example", "subject-123")
	if err != nil {
		t.Fatalf("oidcIdentityClaimID: %v", err)
	}
	if strings.Contains(oidcClaimID, "issuer") || strings.Contains(oidcClaimID, "/") {
		t.Fatalf("OIDC claim ID %q should not expose the issuer or subject", oidcClaimID)
	}
}

func TestIdentityClaimIDsReturnActiveKeyFirstForRotation(t *testing.T) {
	core, _ := setupTestCore(t)
	core.config.IdentityClaims = testIdentityClaimsConfig()

	ids, err := core.loginIdentityClaimIDs("alice")
	if err != nil {
		t.Fatalf("loginIdentityClaimIDs: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("claim IDs = %d, want active plus old key", len(ids))
	}
	if !strings.HasPrefix(ids[0], "login_v2_") {
		t.Fatalf("first claim ID = %q, want active v2 key", ids[0])
	}
	if !strings.HasPrefix(ids[1], "login_v1_") {
		t.Fatalf("second claim ID = %q, want old v1 key", ids[1])
	}
}

func TestIdentityClaimClaimReleaseAndReclaim(t *testing.T) {
	core, _ := setupTestCore(t)
	core.config.IdentityClaims = testIdentityClaimsConfig()
	ctx := testContext(t)

	claimID, err := core.loginIdentityClaimID("alice")
	if err != nil {
		t.Fatalf("loginIdentityClaimID: %v", err)
	}
	kind := corev1.IdentityClaimKind_IDENTITY_CLAIM_KIND_LOGIN

	seqClaimed, err := core.claimIdentity(ctx, "U1", "U1", claimID, kind)
	if err != nil {
		t.Fatalf("claimIdentity U1: %v", err)
	}

	seqNoop, err := core.claimIdentity(ctx, "U1", "U1", claimID, kind)
	if err != nil {
		t.Fatalf("claimIdentity U1 idempotent: %v", err)
	}
	if seqNoop != seqClaimed {
		t.Fatalf("idempotent claim seq = %d, want existing seq %d", seqNoop, seqClaimed)
	}

	_, err = core.claimIdentity(ctx, "U2", "U2", claimID, kind)
	if !errors.Is(err, ErrIdentityClaimTaken) {
		t.Fatalf("claimIdentity U2 while active: got %v, want ErrIdentityClaimTaken", err)
	}

	seqReleased, err := core.releaseIdentity(ctx, "U1", "U1", claimID, kind)
	if err != nil {
		t.Fatalf("releaseIdentity U1: %v", err)
	}
	if seqReleased <= seqClaimed {
		t.Fatalf("release seq = %d, want > claim seq %d", seqReleased, seqClaimed)
	}

	seqReclaimed, err := core.claimIdentity(ctx, "U2", "U2", claimID, kind)
	if err != nil {
		t.Fatalf("claimIdentity U2 after release: %v", err)
	}
	if seqReclaimed <= seqReleased {
		t.Fatalf("reclaim seq = %d, want > release seq %d", seqReclaimed, seqReleased)
	}

	state, err := core.readIdentityClaimState(ctx, claimID)
	if err != nil {
		t.Fatalf("readIdentityClaimState: %v", err)
	}
	if !state.active || state.ownerUserID != "U2" || state.kind != kind {
		t.Fatalf("state = %+v, want active U2 login claim", state)
	}

	published, _, err := core.EventPublisher.SubjectEvents(ctx, events.IdentityClaimAggregate(claimID).AllEventsFilter())
	if err != nil {
		t.Fatalf("SubjectEvents: %v", err)
	}
	if len(published) != 3 {
		t.Fatalf("published events = %d, want claim/release/reclaim", len(published))
	}
}

func testIdentityClaimsConfig() config.IdentityClaimsConfig {
	return config.IdentityClaimsConfig{
		ActiveKeyID: "v2",
		Keys: []config.IdentityClaimKeyConfig{{
			ID:     "v1",
			Secret: "old-identity-claim-secret",
		}, {
			ID:     "v2",
			Secret: "active-identity-claim-secret",
		}},
	}
}
