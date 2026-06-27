package http_server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"hmans.de/chatto/internal/config"
	"hmans.de/chatto/internal/core"
	"hmans.de/chatto/internal/core/linkpreview"
	graphauth "hmans.de/chatto/internal/graph/auth"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

const (
	oidcAvatarFetchTimeout = 10 * time.Second
	oidcAvatarMaxBytes     = 5 * 1024 * 1024
)

var oidcAvatarClient = linkpreview.NewSSRFSafeClient(oidcAvatarFetchTimeout)

// oidcProvider holds the lazily-initialized OIDC provider, oauth2 config, and
// token verifier. Initialized on first login attempt. Retries on failure.
type oidcProvider struct {
	mu           sync.Mutex
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	ready        bool
}

func (o *oidcProvider) init(issuerURL, clientID, clientSecret, redirectURL string, scopes []string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.ready {
		return nil
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		log.Error("Failed to initialize OIDC provider", "issuer", issuerURL, "error", err)
		return err
	}

	o.provider = provider
	o.oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       append([]string(nil), scopes...),
	}
	o.verifier = provider.Verifier(&oidc.Config{ClientID: clientID})
	o.ready = true

	log.Info("OIDC provider initialized", "issuer", issuerURL)
	return nil
}

func (s *HTTPServer) setupOIDCRoutes() {
	providers := s.config.Auth.Providers
	if len(providers) == 0 {
		return
	}

	configured := make(map[string]*authProviderRuntime, len(providers))
	for _, providerConfig := range providers {
		runtime, err := newAuthProviderRuntime(providerConfig, s.providerCallbackURL(providerConfig.ID))
		if err != nil {
			s.logger.Error("Skipping invalid auth provider", "provider_id", providerConfig.ID, "provider_type", providerConfig.Type, "error", err)
			continue
		}
		configured[providerConfig.ID] = runtime
	}
	if len(configured) == 0 {
		return
	}

	auth := s.router.Group("/auth")
	auth.Use(func(c *gin.Context) {
		s.requestContextWithAuditMetadata(c)
		c.Next()
	})

	auth.GET("providers/:providerID", func(c *gin.Context) {
		providerRuntime := configured[c.Param("providerID")]
		if providerRuntime == nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_not_found")
			return
		}

		s.handleProviderStart(c, providerRuntime)
	})

	auth.GET("providers/:providerID/callback", func(c *gin.Context) {
		providerRuntime := configured[c.Param("providerID")]
		if providerRuntime == nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_not_found")
			return
		}

		s.handleProviderCallback(c, providerRuntime)
	})
	if legacyRuntime := s.legacyOIDCRuntime(configured); legacyRuntime != nil {
		auth.GET("oidc", func(c *gin.Context) {
			s.handleProviderStart(c, legacyRuntime)
		})
		auth.GET("oidc/callback", func(c *gin.Context) {
			s.handleProviderCallback(c, legacyRuntime)
		})
	}

	auth.GET("pending-oidc/:token", s.handlePendingOIDCGet)
	auth.POST("pending-oidc/:token/create", s.handlePendingOIDCCreate)
	auth.POST("pending-oidc/:token/link-password", s.handlePendingOIDCLinkPassword)
	auth.POST("pending-oidc/:token/link-current", s.handlePendingOIDCLinkCurrent)
	auth.POST("pending-oidc/:token/cancel", s.handlePendingOIDCCancel)
}

func (s *HTTPServer) handleProviderStart(c *gin.Context, providerRuntime *authProviderRuntime) {
	session := sessions.Default(c)

	// Store redirect URL if provided
	if redirect := c.Query("redirect"); redirect != "" {
		if isValidInternalRedirect(redirect) {
			session.Set("oauth_redirect", redirect)
		}
	}
	if c.Query("mode") == "link" {
		session.Set(providerSessionKey(providerRuntime.config.ID, "mode"), "link")
		if linked := s.authenticatedUser(c); linked != nil {
			session.Set(providerSessionKey(providerRuntime.config.ID, "link_user_id"), linked.Id)
		}
	} else {
		session.Delete(providerSessionKey(providerRuntime.config.ID, "mode"))
		session.Delete(providerSessionKey(providerRuntime.config.ID, "link_user_id"))
	}

	state, err := randomString(32)
	if err != nil {
		log.Error("Failed to generate provider auth state", "provider_id", providerRuntime.config.ID, "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}

	session.Set(providerSessionKey(providerRuntime.config.ID, "state"), state)

	var authURL string
	if !providerRuntime.ensureOIDC(c) {
		return
	}
	codeVerifier, err := randomString(64)
	if err != nil {
		log.Error("Failed to generate PKCE code verifier", "provider_id", providerRuntime.config.ID, "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}
	session.Set(providerSessionKey(providerRuntime.config.ID, "code_verifier"), codeVerifier)
	codeChallenge := s256Challenge(codeVerifier)
	authURL = providerRuntime.oidc.oauth2Config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	if err := session.Save(); err != nil {
		log.Error("Failed to save provider auth session", "provider_id", providerRuntime.config.ID, "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (s *HTTPServer) handleProviderCallback(c *gin.Context, providerRuntime *authProviderRuntime) {
	session := sessions.Default(c)
	ctx := c.Request.Context()

	// Verify state
	expectedState, _ := session.Get(providerSessionKey(providerRuntime.config.ID, "state")).(string)
	if expectedState == "" || c.Query("state") != expectedState {
		log.Warn("Provider callback state mismatch", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type)
		session.Delete(providerSessionKey(providerRuntime.config.ID, "state"))
		session.Delete(providerSessionKey(providerRuntime.config.ID, "code_verifier"))
		session.Delete(providerSessionKey(providerRuntime.config.ID, "mode"))
		session.Delete(providerSessionKey(providerRuntime.config.ID, "link_user_id"))
		_ = session.Save()
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}

	session.Delete(providerSessionKey(providerRuntime.config.ID, "state"))
	mode, _ := session.Get(providerSessionKey(providerRuntime.config.ID, "mode")).(string)
	session.Delete(providerSessionKey(providerRuntime.config.ID, "mode"))
	linkUserID, _ := session.Get(providerSessionKey(providerRuntime.config.ID, "link_user_id")).(string)
	session.Delete(providerSessionKey(providerRuntime.config.ID, "link_user_id"))

	// Check for error from provider
	if errCode := c.Query("error"); errCode != "" {
		log.Warn("Provider returned auth error", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type, "error", errCode)
		session.Delete(providerSessionKey(providerRuntime.config.ID, "code_verifier"))
		_ = session.Save()
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_denied")
		return
	}

	if !providerRuntime.ensureOIDC(c) {
		return
	}

	identity, err := providerRuntime.resolveIdentity(c, session)
	if err != nil {
		log.Error("Provider callback failed", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type, "error", err)
		session.Delete(providerSessionKey(providerRuntime.config.ID, "code_verifier"))
		_ = session.Save()
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}
	session.Delete(providerSessionKey(providerRuntime.config.ID, "code_verifier"))
	_ = session.Save()

	user, err := s.core.GetUserByExternalIdentity(ctx, identity.issuer, identity.subject)
	if err != nil {
		log.Error("Failed to lookup user by external identity", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type, "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}
	if user == nil {
		if mode == "link" {
			if linked := s.authenticatedUser(c); linked != nil {
				if linkUserID == "" || linked.Id != linkUserID {
					c.Redirect(http.StatusTemporaryRedirect, "/chat?error=oidc_link_failed")
					return
				}
				if _, err := s.core.LinkOIDCIdentityToUser(ctx, linked.Id, identity.toProvisionProfile(providerRuntime.config)); err != nil {
					log.Error("Failed to link OIDC identity", "provider_id", providerRuntime.config.ID, "userId", linked.Id, "error", err)
					c.Redirect(http.StatusTemporaryRedirect, "/chat?error=oidc_link_failed")
					return
				}
				c.Redirect(http.StatusTemporaryRedirect, s.providerRedirectURL(session, "/chat?oidc_linked=1"))
				return
			}
		}
		log.Info("Provider login has no linked Chatto account", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type)
		token, err := s.createPendingOIDCIdentity(c, session, providerRuntime.config, identity, mode, linkUserID)
		if err != nil {
			log.Error("Failed to create pending OIDC identity", "provider_id", providerRuntime.config.ID, "error", err)
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/login/oidc?token="+url.QueryEscape(token))
		return
	}

	if mode == "link" {
		if linked := s.authenticatedUser(c); linked != nil {
			if linked.Id == user.Id {
				c.Redirect(http.StatusTemporaryRedirect, s.providerRedirectURL(session, "/chat?oidc_linked=1"))
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, "/chat?error=oidc_identity_conflict")
			return
		}
	}

	log.Info("Provider login matched by external identity", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type, "userId", user.Id)

	if identity.avatarURL != "" {
		existingAvatar, _ := s.core.GetUserAvatar(ctx, user.Id)
		if existingAvatar == nil {
			if err := fetchAndUploadAvatarFromURL(ctx, identity.avatarURL, s, user.Id); err != nil {
				log.Error("Failed to fetch provider avatar", "provider_id", providerRuntime.config.ID, "error", err)
			}
		}
	}

	if err := s.completeProviderLogin(c, session, user.Id, providerRuntime.config); err != nil {
		log.Error("Failed to complete provider login", "provider_id", providerRuntime.config.ID, "provider_type", providerRuntime.config.Type, "userId", user.Id, "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return
	}
}

type authProviderRuntime struct {
	config      config.AuthProviderConfig
	callbackURL string
	oidc        *oidcProvider
}

type resolvedProviderIdentity struct {
	issuer        string
	subject       string
	email         string
	emailVerified bool
	name          string
	username      string
	avatarURL     string
}

func newAuthProviderRuntime(providerConfig config.AuthProviderConfig, callbackURL string) (*authProviderRuntime, error) {
	runtime := &authProviderRuntime{config: providerConfig, callbackURL: callbackURL}
	switch providerConfig.Type {
	case config.AuthProviderTypeOpenIDConnect:
		runtime.oidc = &oidcProvider{}
	default:
		return nil, fmt.Errorf("unsupported auth provider type %q", providerConfig.Type)
	}
	return runtime, nil
}

func (s *HTTPServer) legacyOIDCRuntime(configured map[string]*authProviderRuntime) *authProviderRuntime {
	if len(configured) != 1 {
		return nil
	}
	for _, runtime := range configured {
		legacyRuntime, err := newAuthProviderRuntime(runtime.config, s.legacyOIDCCallbackURL())
		if err != nil {
			s.logger.Error("Skipping legacy OIDC alias runtime", "provider_id", runtime.config.ID, "error", err)
			return nil
		}
		return legacyRuntime
	}
	return nil
}

func (s *HTTPServer) legacyOIDCCallbackURL() string {
	baseURL := strings.TrimRight(s.config.Webserver.URL, "/")
	return baseURL + "/auth/oidc/callback"
}

func providerScopes(providerConfig config.AuthProviderConfig) []string {
	if len(providerConfig.Scopes) > 0 {
		scopes := append([]string(nil), providerConfig.Scopes...)
		if providerConfig.Type == config.AuthProviderTypeOpenIDConnect && !hasScope(scopes, oidc.ScopeOpenID) {
			scopes = append([]string{oidc.ScopeOpenID}, scopes...)
		}
		return scopes
	}
	if providerConfig.Type == config.AuthProviderTypeOpenIDConnect {
		scopes := []string{oidc.ScopeOpenID, "profile"}
		if providerConfig.RequestEmailOrDefault() {
			scopes = append(scopes, "email")
		}
		return scopes
	}
	return nil
}

func hasScope(scopes []string, target string) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func (s *HTTPServer) providerCallbackURL(providerID string) string {
	return strings.TrimRight(s.config.Webserver.URL, "/") + "/auth/providers/" + url.PathEscape(providerID) + "/callback"
}

func providerSessionKey(providerID, name string) string {
	return "provider_" + providerID + "_" + name
}

func (r *authProviderRuntime) ensureOIDC(c *gin.Context) bool {
	if r.oidc == nil {
		return true
	}
	if err := r.oidc.init(r.config.IssuerURL, r.config.ClientID, r.config.ClientSecret, r.callbackURL, providerScopes(r.config)); err != nil {
		log.Error("OIDC provider not available", "provider_id", r.config.ID, "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=provider_failed")
		return false
	}
	return true
}

func (r *authProviderRuntime) resolveIdentity(c *gin.Context, session sessions.Session) (resolvedProviderIdentity, error) {
	return r.resolveOIDCIdentity(c, session)
}

func (r *authProviderRuntime) resolveOIDCIdentity(c *gin.Context, session sessions.Session) (resolvedProviderIdentity, error) {
	ctx := c.Request.Context()
	codeVerifier, _ := session.Get(providerSessionKey(r.config.ID, "code_verifier")).(string)
	if codeVerifier == "" {
		return resolvedProviderIdentity{}, fmt.Errorf("missing code verifier")
	}

	// Exchange authorization code for tokens
	token, err := r.oidc.oauth2Config.Exchange(ctx, c.Query("code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return resolvedProviderIdentity{}, fmt.Errorf("token exchange failed: %w", err)
	}

	// Extract and verify the ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return resolvedProviderIdentity{}, fmt.Errorf("token response missing id_token")
	}

	idToken, err := r.oidc.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return resolvedProviderIdentity{}, fmt.Errorf("id token verification failed: %w", err)
	}

	// Extract claims from the ID token first
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		PreferredUser string `json:"preferred_username"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return resolvedProviderIdentity{}, fmt.Errorf("parse id token claims: %w", err)
	}

	log.Info("OIDC token verified", "provider_id", r.config.ID, "issuer", idToken.Issuer)

	// Some providers (e.g. Zitadel) don't include email in the ID token.
	// Fall back to the userinfo endpoint.
	if claims.Email == "" {
		log.Info("OIDC ID token missing email, falling back to userinfo", "provider_id", r.config.ID)
		userInfo, err := r.oidc.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
		if err != nil {
			return resolvedProviderIdentity{}, fmt.Errorf("fetch userinfo: %w", err)
		}
		if err := userInfo.Claims(&claims); err != nil {
			return resolvedProviderIdentity{}, fmt.Errorf("parse userinfo claims: %w", err)
		}
	}

	if claims.Email != "" && claims.EmailVerified {
		claims.Email = strings.ToLower(strings.TrimSpace(claims.Email))
	}
	return resolvedProviderIdentity{
		issuer:        idToken.Issuer,
		subject:       idToken.Subject,
		email:         claims.Email,
		emailVerified: claims.EmailVerified,
		name:          claims.Name,
		username:      claims.PreferredUser,
		avatarURL:     claims.Picture,
	}, nil
}

func (identity resolvedProviderIdentity) toProvisionProfile(providerConfig config.AuthProviderConfig) core.OIDCProvisionProfile {
	return core.OIDCProvisionProfile{
		ProviderID:    providerConfig.ID,
		ProviderLabel: providerConfig.LabelOrDefault(),
		Issuer:        identity.issuer,
		Subject:       identity.subject,
		Email:         identity.email,
		EmailVerified: identity.emailVerified,
		Name:          identity.name,
		Username:      identity.username,
	}
}

func pendingOIDCToProvisionProfile(pending *core.PendingOIDCIdentity) core.OIDCProvisionProfile {
	return core.OIDCProvisionProfile{
		ProviderID:    pending.ProviderID,
		ProviderLabel: pending.ProviderLabel,
		Issuer:        pending.Issuer,
		Subject:       pending.Subject,
		Email:         pending.Email,
		EmailVerified: pending.EmailVerified,
		Name:          pending.Name,
		Username:      pending.Username,
	}
}

func (s *HTTPServer) createPendingOIDCIdentity(c *gin.Context, session sessions.Session, providerConfig config.AuthProviderConfig, identity resolvedProviderIdentity, mode, linkUserID string) (string, error) {
	redirectURL := "/"
	if redirect := session.Get("oauth_redirect"); redirect != nil {
		if r, ok := redirect.(string); ok && r != "" && isValidInternalRedirect(r) {
			redirectURL = r
		}
		session.Delete("oauth_redirect")
		_ = session.Save()
	}
	return s.core.CreatePendingOIDCIdentity(c.Request.Context(), core.PendingOIDCIdentity{
		ProviderID:    providerConfig.ID,
		ProviderLabel: providerConfig.LabelOrDefault(),
		Issuer:        identity.issuer,
		Subject:       identity.subject,
		Email:         identity.email,
		EmailVerified: identity.emailVerified,
		Name:          identity.name,
		Username:      identity.username,
		AvatarURL:     identity.avatarURL,
		RedirectURL:   redirectURL,
		Mode:          mode,
		LinkUserID:    linkUserID,
	})
}

func (s *HTTPServer) providerRedirectURL(session sessions.Session, fallback string) string {
	if redirect := session.Get("oauth_redirect"); redirect != nil {
		if r, ok := redirect.(string); ok && r != "" && isValidInternalRedirect(r) {
			session.Delete("oauth_redirect")
			_ = session.Save()
			return r
		}
	}
	return fallback
}

func (s *HTTPServer) handlePendingOIDCGet(c *gin.Context) {
	pending, err := s.core.GetPendingOIDCIdentity(c.Request.Context(), c.Param("token"))
	if err != nil {
		status := http.StatusInternalServerError
		code := "provider_failed"
		if errors.Is(err, core.ErrPendingOIDCNotFound) || errors.Is(err, core.ErrPendingOIDCExpired) {
			status = http.StatusNotFound
			code = "pending_oidc_not_found"
		}
		c.JSON(status, gin.H{"error": code})
		return
	}
	canLinkCurrent := false
	if pending.LinkUserID != "" {
		if user := s.authenticatedUser(c); user != nil && user.Id == pending.LinkUserID {
			canLinkCurrent = true
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"providerId":     pending.ProviderID,
		"providerLabel":  pending.ProviderLabel,
		"email":          pending.Email,
		"emailVerified":  pending.EmailVerified,
		"name":           pending.Name,
		"username":       pending.Username,
		"mode":           pending.Mode,
		"canLinkCurrent": canLinkCurrent,
	})
}

func (s *HTTPServer) handlePendingOIDCCreate(c *gin.Context) {
	session := sessions.Default(c)
	ctx := c.Request.Context()
	token := c.Param("token")
	pending, err := s.core.GetPendingOIDCIdentity(ctx, token)
	if err != nil {
		s.pendingOIDCError(c, err)
		return
	}
	if user, err := s.core.GetUserByExternalIdentity(ctx, pending.Issuer, pending.Subject); err != nil {
		log.Error("Failed to lookup pending OIDC identity", "provider_id", pending.ProviderID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
		return
	} else if user != nil {
		_ = s.core.DeletePendingOIDCIdentity(ctx, token)
		setPendingOIDCRedirect(session, pending)
		if err := s.completeProviderLogin(c, session, user.Id, config.AuthProviderConfig{ID: pending.ProviderID, Type: config.AuthProviderTypeOpenIDConnect}); err != nil {
			log.Error("Failed to complete linked pending OIDC login", "provider_id", pending.ProviderID, "userId", user.Id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
		}
		return
	}
	user, _, err := s.core.ProvisionOIDCUser(ctx, pendingOIDCToProvisionProfile(pending))
	if err != nil {
		log.Error("Failed to provision OIDC user", "provider_id", pending.ProviderID, "error", err)
		c.JSON(http.StatusConflict, gin.H{"error": "oidc_provision_failed"})
		return
	}
	_ = s.core.DeletePendingOIDCIdentity(ctx, token)
	if pending.AvatarURL != "" {
		if existingAvatar, _ := s.core.GetUserAvatar(ctx, user.Id); existingAvatar == nil {
			if err := fetchAndUploadAvatarFromURL(ctx, pending.AvatarURL, s, user.Id); err != nil {
				log.Error("Failed to fetch provider avatar", "provider_id", pending.ProviderID, "error", err)
			}
		}
	}
	setPendingOIDCRedirect(session, pending)
	if err := s.completeProviderLogin(c, session, user.Id, config.AuthProviderConfig{ID: pending.ProviderID, Type: config.AuthProviderTypeOpenIDConnect}); err != nil {
		log.Error("Failed to complete provisioned OIDC login", "provider_id", pending.ProviderID, "userId", user.Id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
		return
	}
}

func (s *HTTPServer) handlePendingOIDCLinkPassword(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier"`
		Login      string `json:"login"`
		Password   string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}
	identifier := req.Identifier
	if identifier == "" {
		identifier = req.Login
	}
	if strings.TrimSpace(identifier) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required"})
		return
	}

	ctx := c.Request.Context()
	token := c.Param("token")
	pending, err := s.core.GetPendingOIDCIdentity(ctx, token)
	if err != nil {
		s.pendingOIDCError(c, err)
		return
	}
	user, authGeneration, err := s.core.VerifyPasswordWithAuthGeneration(ctx, identifier, req.Password)
	if err != nil {
		if auditErr := s.core.RecordLoginFailed(ctx, identifier); auditErr != nil {
			log.Warn("Failed to append failed-login audit event", "error", auditErr)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if _, err := s.core.LinkOIDCIdentityToUser(ctx, user.Id, pendingOIDCToProvisionProfile(pending)); err != nil {
		log.Error("Failed to link pending OIDC identity", "provider_id", pending.ProviderID, "userId", user.Id, "error", err)
		c.JSON(http.StatusConflict, gin.H{"error": "oidc_identity_conflict"})
		return
	}
	_ = s.core.DeletePendingOIDCIdentity(ctx, token)
	setPendingOIDCRedirect(sessions.Default(c), pending)
	if err := s.completeProviderLoginForGeneration(c, sessions.Default(c), user.Id, config.AuthProviderConfig{ID: pending.ProviderID, Type: config.AuthProviderTypeOpenIDConnect}, authGeneration); err != nil {
		log.Error("Failed to complete linked OIDC login", "provider_id", pending.ProviderID, "userId", user.Id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
		return
	}
}

func (s *HTTPServer) handlePendingOIDCLinkCurrent(c *gin.Context) {
	user := s.authenticatedUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication_required"})
		return
	}
	ctx := c.Request.Context()
	token := c.Param("token")
	pending, err := s.core.GetPendingOIDCIdentity(ctx, token)
	if err != nil {
		s.pendingOIDCError(c, err)
		return
	}
	if pending.LinkUserID == "" || pending.LinkUserID != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "oidc_link_not_allowed"})
		return
	}
	if _, err := s.core.LinkOIDCIdentityToUser(ctx, user.Id, pendingOIDCToProvisionProfile(pending)); err != nil {
		log.Error("Failed to link pending OIDC identity to current user", "provider_id", pending.ProviderID, "userId", user.Id, "error", err)
		c.JSON(http.StatusConflict, gin.H{"error": "oidc_identity_conflict"})
		return
	}
	_ = s.core.DeletePendingOIDCIdentity(ctx, token)
	setPendingOIDCRedirect(sessions.Default(c), pending)
	if err := s.completeProviderLogin(c, sessions.Default(c), user.Id, config.AuthProviderConfig{ID: pending.ProviderID, Type: config.AuthProviderTypeOpenIDConnect}); err != nil {
		log.Error("Failed to complete current-account OIDC login", "provider_id", pending.ProviderID, "userId", user.Id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
		return
	}
}

func setPendingOIDCRedirect(session sessions.Session, pending *core.PendingOIDCIdentity) {
	if pending == nil {
		return
	}
	if pending.RedirectURL != "" && isValidInternalRedirect(pending.RedirectURL) {
		session.Set("oauth_redirect", pending.RedirectURL)
	}
}

func (s *HTTPServer) handlePendingOIDCCancel(c *gin.Context) {
	if err := s.core.DeletePendingOIDCIdentity(c.Request.Context(), c.Param("token")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *HTTPServer) pendingOIDCError(c *gin.Context, err error) {
	if errors.Is(err, core.ErrPendingOIDCNotFound) || errors.Is(err, core.ErrPendingOIDCExpired) {
		c.JSON(http.StatusNotFound, gin.H{"error": "pending_oidc_not_found"})
		return
	}
	log.Error("Pending OIDC operation failed", "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "provider_failed"})
}

func (s *HTTPServer) authenticatedUser(c *gin.Context) *corev1.User {
	req := s.injectUserIntoContext(c)
	return graphauth.ForContext(req.Context())
}

func (s *HTTPServer) completeProviderLogin(c *gin.Context, session sessions.Session, userID string, providerConfig config.AuthProviderConfig) error {
	source := providerConfig.Type + "_login"
	return s.completeProviderLoginWithSource(c, session, userID, providerConfig, source, nil)
}

func (s *HTTPServer) completeProviderLoginForGeneration(c *gin.Context, session sessions.Session, userID string, providerConfig config.AuthProviderConfig, authGeneration uint64) error {
	source := providerConfig.Type + "_login"
	return s.completeProviderLoginWithSource(c, session, userID, providerConfig, source, &authGeneration)
}

func (s *HTTPServer) completeProviderLoginWithSource(c *gin.Context, session sessions.Session, userID string, providerConfig config.AuthProviderConfig, source string, authGeneration *uint64) error {
	ctx := c.Request.Context()
	if authGeneration != nil {
		if err := s.createCookieSessionForGeneration(c, userID, source, *authGeneration); err != nil {
			return fmt.Errorf("save cookie session: %w", err)
		}
	} else {
		if err := s.createCookieSession(c, userID, source); err != nil {
			return fmt.Errorf("save cookie session: %w", err)
		}
	}
	if err := s.ensureCSRFToken(c); err != nil {
		session = sessions.Default(c)
		cookieUserID, cookieSessionID, _ := cookieSessionIDs(session)
		_ = s.core.RevokeCookieSession(ctx, cookieUserID, cookieSessionID)
		session.Clear()
		_ = session.Save()
		clearCSRFCookie(c)
		return fmt.Errorf("create csrf token: %w", err)
	}
	if err := s.core.RecordLoginSucceeded(ctx, userID, providerConfig.Type+":"+providerConfig.ID); err != nil {
		session = sessions.Default(c)
		cookieUserID, cookieSessionID, _ := cookieSessionIDs(session)
		_ = s.core.RevokeCookieSession(ctx, cookieUserID, cookieSessionID)
		session.Clear()
		_ = session.Save()
		clearCSRFCookie(c)
		return fmt.Errorf("append login audit event: %w", err)
	}

	if hasPendingOAuthAuthorize(session) {
		authGeneration, err := s.core.CurrentAuthGeneration(ctx, userID)
		if err != nil {
			return fmt.Errorf("read auth generation for OAuth authorize: %w", err)
		}
		if wantsJSONResponse(c) {
			redirectURL, ok := s.pendingOAuthAuthorizeRedirectURL(c, userID, authGeneration)
			if !ok {
				return nil
			}
			c.JSON(http.StatusOK, gin.H{"redirectUrl": redirectURL})
			return nil
		}
		s.continueOAuthAuthorize(c, userID, authGeneration)
		return nil
	}

	redirectURL := "/"
	if redirect := session.Get("oauth_redirect"); redirect != nil {
		if r, ok := redirect.(string); ok && r != "" && isValidInternalRedirect(r) {
			redirectURL = r
		}
		session.Delete("oauth_redirect")
		_ = session.Save()
	}

	var bearerToken string
	var err error
	if authGeneration != nil {
		bearerToken, err = s.core.CreateAuthTokenWithSourceGeneration(ctx, userID, source, *authGeneration)
	} else {
		bearerToken, err = s.core.CreateAuthTokenWithSource(ctx, userID, source)
	}
	if err == nil {
		separator := "?"
		if strings.Contains(redirectURL, "?") {
			separator = "&"
		}
		redirectURL = redirectURL + separator + "token=" + bearerToken
	} else {
		log.Warn("Failed to create auth token on provider login", "provider_id", providerConfig.ID, "userId", userID, "error", err)
	}

	if wantsJSONResponse(c) {
		c.JSON(http.StatusOK, gin.H{"redirectUrl": redirectURL})
		return nil
	}
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	return nil
}

func wantsJSONResponse(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Accept"), "application/json")
}

func (s *HTTPServer) pendingOAuthAuthorizeRedirectURL(c *gin.Context, userID string, authGeneration uint64) (string, bool) {
	params, err := readPendingOAuthAuthorize(sessions.Default(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "No pending authorization request",
		})
		return "", false
	}
	redirectOrigin, ok := s.allowedOAuthRedirectOrigin(params.RedirectURI)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Invalid redirect_uri",
		})
		return "", false
	}
	consented, err := s.core.HasOAuthConsent(c.Request.Context(), userID, redirectOrigin)
	if err != nil {
		log.Error("Failed to check OAuth consent", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to check OAuth consent",
		})
		return "", false
	}
	if !consented {
		return "/oauth/consent", true
	}
	return s.completeOAuthAuthorizeURL(c, userID, authGeneration)
}

// randomString generates a URL-safe random string of n bytes.
func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// s256Challenge computes the S256 PKCE code challenge from a code verifier.
func s256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// fetchAndUploadAvatarFromURL downloads an avatar from a URL and uploads it.
func fetchAndUploadAvatarFromURL(ctx context.Context, avatarURL string, s *HTTPServer, userID string) error {
	// Validate the URL before making a request
	u, err := url.Parse(avatarURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil // silently skip invalid URLs
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return err
	}

	resp, err := oidcAvatarClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	avatarData, err := io.ReadAll(io.LimitReader(resp.Body, oidcAvatarMaxBytes+1))
	if err != nil {
		return err
	}
	if len(avatarData) > oidcAvatarMaxBytes {
		return fmt.Errorf("avatar exceeds maximum size of %d bytes", oidcAvatarMaxBytes)
	}

	asset, err := s.core.UploadUserAvatar(ctx, userID, bytes.NewReader(avatarData))
	if err != nil {
		return err
	}

	return s.core.SetUserAvatar(ctx, userID, asset)
}
