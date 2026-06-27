package http_server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"hmans.de/chatto/internal/config"
	"hmans.de/chatto/internal/core"
)

func TestProviderScopesForOIDC(t *testing.T) {
	t.Run("default requests openid profile email", func(t *testing.T) {
		scopes := providerScopes(config.AuthProviderConfig{Type: config.AuthProviderTypeOpenIDConnect})
		want := []string{oidc.ScopeOpenID, "profile", "email"}
		if !slices.Equal(scopes, want) {
			t.Fatalf("providerScopes() = %v, want %v", scopes, want)
		}
	})

	t.Run("request_email false keeps openid profile", func(t *testing.T) {
		requestEmail := false
		scopes := providerScopes(config.AuthProviderConfig{
			Type:         config.AuthProviderTypeOpenIDConnect,
			RequestEmail: &requestEmail,
		})
		want := []string{oidc.ScopeOpenID, "profile"}
		if !slices.Equal(scopes, want) {
			t.Fatalf("providerScopes() = %v, want %v", scopes, want)
		}
	})

	t.Run("custom scopes are honored with openid required", func(t *testing.T) {
		scopes := providerScopes(config.AuthProviderConfig{
			Type:   config.AuthProviderTypeOpenIDConnect,
			Scopes: []string{"groups", "profile"},
		})
		want := []string{oidc.ScopeOpenID, "groups", "profile"}
		if !slices.Equal(scopes, want) {
			t.Fatalf("providerScopes() = %v, want %v", scopes, want)
		}
	})
}

func TestOIDCProviderRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var issuer *httptest.Server
	issuer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 issuer.URL,
			"authorization_endpoint": issuer.URL + "/authorize",
			"token_endpoint":         issuer.URL + "/token",
			"jwks_uri":               issuer.URL + "/keys",
			"userinfo_endpoint":      issuer.URL + "/userinfo",
		})
	}))
	t.Cleanup(issuer.Close)

	router := gin.New()
	sessionStore := cookie.NewStore([]byte("test-secret-key-32-bytes-long!!"))
	router.Use(sessions.Sessions("chatto_session", sessionStore))

	s := &HTTPServer{
		config: config.ChattoConfig{
			Webserver: config.WebserverConfig{
				URL: "http://chat.example",
			},
			Auth: config.AuthConfig{
				Providers: []config.AuthProviderConfig{{
					ID:           "hub",
					Type:         config.AuthProviderTypeOpenIDConnect,
					IssuerURL:    issuer.URL,
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				}},
			},
		},
		router: router,
		logger: log.WithPrefix("test.HTTP"),
	}
	s.setupOIDCRoutes()

	t.Run("legacy login route redirects with legacy callback URI", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/oidc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Fatalf("GET /auth/oidc status = %d, want %d", w.Code, http.StatusTemporaryRedirect)
		}
		assertRedirectURI(t, w.Header().Get("Location"), "http://chat.example/auth/oidc/callback")
	})

	t.Run("provider login route keeps provider callback URI", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/providers/hub", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Fatalf("GET /auth/providers/hub status = %d, want %d", w.Code, http.StatusTemporaryRedirect)
		}
		assertRedirectURI(t, w.Header().Get("Location"), "http://chat.example/auth/providers/hub/callback")
	})

	t.Run("legacy callback route remains registered", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/oidc/callback?state=missing", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Fatalf("GET /auth/oidc/callback status = %d, want %d", w.Code, http.StatusTemporaryRedirect)
		}
		if location := w.Header().Get("Location"); location != "/login?error=provider_failed" {
			t.Fatalf("legacy callback Location = %q, want provider_failed redirect", location)
		}
	})
}

func assertRedirectURI(t *testing.T, location, want string) {
	t.Helper()
	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("redirect Location %q did not parse: %v", location, err)
	}
	got := redirectURL.Query().Get("redirect_uri")
	if got != want {
		t.Fatalf("redirect_uri = %q, want %q; Location = %q", got, want, location)
	}
}

func TestPendingOIDCRoutes(t *testing.T) {
	ts, client, chattoCore := setupTestHTTPServerWithHook(t, func(s *HTTPServer) {
		s.config.Auth.Providers = []config.AuthProviderConfig{{
			ID:           "hub",
			Type:         config.AuthProviderTypeOpenIDConnect,
			Label:        "Chatto Hub",
			IssuerURL:    "https://issuer.example",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		}}
		s.setupOIDCRoutes()
	})

	createPendingForLinkUser := func(t *testing.T, subject, redirect, linkUserID string) string {
		t.Helper()
		token, err := chattoCore.CreatePendingOIDCIdentity(t.Context(), core.PendingOIDCIdentity{
			ProviderID:    "hub",
			ProviderLabel: "Chatto Hub",
			Issuer:        "https://issuer.example",
			Subject:       subject,
			Name:          "OIDC User",
			Username:      "oidc-user",
			RedirectURL:   redirect,
			LinkUserID:    linkUserID,
		})
		if err != nil {
			t.Fatalf("CreatePendingOIDCIdentity: %v", err)
		}
		return token
	}
	createPending := func(t *testing.T, subject, redirect string) string {
		t.Helper()
		return createPendingForLinkUser(t, subject, redirect, "")
	}

	t.Run("get exposes safe pending metadata", func(t *testing.T) {
		token := createPending(t, "subject-get", "/chat/abc/settings/account")

		resp := doPendingOIDCRequest(t, client, http.MethodGet, ts.URL+"/auth/pending-oidc/"+url.PathEscape(token), nil)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET pending status = %d, want 200: %s", resp.StatusCode, readBody(t, resp))
		}
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode pending response: %v", err)
		}
		if body["providerId"] != "hub" || body["providerLabel"] != "Chatto Hub" || body["username"] != "oidc-user" {
			t.Fatalf("pending response = %#v", body)
		}
		if _, ok := body["subject"]; ok {
			t.Fatalf("pending response leaked subject: %#v", body)
		}
	})

	t.Run("create provisions a passwordless user and preserves redirect", func(t *testing.T) {
		token := createPending(t, "subject-create", "/chat/origin/settings/account")

		resp := doPendingOIDCRequest(t, client, http.MethodPost, ts.URL+"/auth/pending-oidc/"+url.PathEscape(token)+"/create", nil)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST create status = %d, want 200: %s", resp.StatusCode, readBody(t, resp))
		}
		redirectURL := decodeRedirectURL(t, resp)
		if parsed, err := url.Parse(redirectURL); err != nil || parsed.Path != "/chat/origin/settings/account" || parsed.Query().Get("token") == "" {
			t.Fatalf("redirectUrl = %q, want account settings path with bearer token", redirectURL)
		}
		if _, err := chattoCore.GetPendingOIDCIdentity(t.Context(), token); err == nil {
			t.Fatal("pending token still exists after create")
		}
		user, err := chattoCore.GetUserByExternalIdentity(t.Context(), "https://issuer.example", "subject-create")
		if err != nil {
			t.Fatalf("GetUserByExternalIdentity: %v", err)
		}
		if user == nil || user.Login != "oidc-user" {
			t.Fatalf("linked user = %v, want provisioned oidc-user", user)
		}
	})

	t.Run("link password attaches identity and preserves redirect", func(t *testing.T) {
		user, err := chattoCore.CreateUser(t.Context(), core.SystemActorID, "local-oidc-user", "Local User", "password123")
		if err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		token := createPending(t, "subject-link-password", "/chat/origin/settings/account")
		payload := []byte(`{"identifier":"local-oidc-user","password":"password123"}`)

		resp := doPendingOIDCRequest(t, client, http.MethodPost, ts.URL+"/auth/pending-oidc/"+url.PathEscape(token)+"/link-password", payload)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST link-password status = %d, want 200: %s", resp.StatusCode, readBody(t, resp))
		}
		redirectURL := decodeRedirectURL(t, resp)
		if parsed, err := url.Parse(redirectURL); err != nil || parsed.Path != "/chat/origin/settings/account" || parsed.Query().Get("token") == "" {
			t.Fatalf("redirectUrl = %q, want account settings path with bearer token", redirectURL)
		}
		linked, err := chattoCore.GetUserByExternalIdentity(t.Context(), "https://issuer.example", "subject-link-password")
		if err != nil {
			t.Fatalf("GetUserByExternalIdentity: %v", err)
		}
		if linked == nil || linked.Id != user.Id {
			t.Fatalf("linked user = %v, want %s", linked, user.Id)
		}
	})

	t.Run("link current attaches identity with bearer token and preserves redirect", func(t *testing.T) {
		user, err := chattoCore.CreateUser(t.Context(), core.SystemActorID, "current-oidc-user", "Current User", "password123")
		if err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		bearerToken, err := chattoCore.CreateAuthTokenWithSource(t.Context(), user.Id, "test")
		if err != nil {
			t.Fatalf("CreateAuthTokenWithSource: %v", err)
		}
		token := createPendingForLinkUser(t, "subject-link-current", "/chat/origin/settings/account", user.Id)

		getResp := doPendingOIDCRequestWithHeaders(
			t,
			client,
			http.MethodGet,
			ts.URL+"/auth/pending-oidc/"+url.PathEscape(token),
			nil,
			map[string]string{"Authorization": "Bearer " + bearerToken},
		)
		defer getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("GET bound pending status = %d, want 200: %s", getResp.StatusCode, readBody(t, getResp))
		}
		var metadata map[string]any
		if err := json.NewDecoder(getResp.Body).Decode(&metadata); err != nil {
			t.Fatalf("decode bound pending metadata: %v", err)
		}
		if metadata["canLinkCurrent"] != true {
			t.Fatalf("canLinkCurrent = %#v, want true", metadata["canLinkCurrent"])
		}

		resp := doPendingOIDCRequestWithHeaders(
			t,
			client,
			http.MethodPost,
			ts.URL+"/auth/pending-oidc/"+url.PathEscape(token)+"/link-current",
			nil,
			map[string]string{"Authorization": "Bearer " + bearerToken},
		)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST link-current status = %d, want 200: %s", resp.StatusCode, readBody(t, resp))
		}
		redirectURL := decodeRedirectURL(t, resp)
		if parsed, err := url.Parse(redirectURL); err != nil || parsed.Path != "/chat/origin/settings/account" || parsed.Query().Get("token") == "" {
			t.Fatalf("redirectUrl = %q, want account settings path with bearer token", redirectURL)
		}
		linked, err := chattoCore.GetUserByExternalIdentity(t.Context(), "https://issuer.example", "subject-link-current")
		if err != nil {
			t.Fatalf("GetUserByExternalIdentity: %v", err)
		}
		if linked == nil || linked.Id != user.Id {
			t.Fatalf("linked user = %v, want %s", linked, user.Id)
		}
		if _, err := chattoCore.GetPendingOIDCIdentity(t.Context(), token); err == nil {
			t.Fatal("pending token still exists after link-current")
		}
	})

	t.Run("link current rejects generic pending token", func(t *testing.T) {
		user, err := chattoCore.CreateUser(t.Context(), core.SystemActorID, "generic-link-victim", "Generic Link Victim", "password123")
		if err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		bearerToken, err := chattoCore.CreateAuthTokenWithSource(t.Context(), user.Id, "test")
		if err != nil {
			t.Fatalf("CreateAuthTokenWithSource: %v", err)
		}
		token := createPending(t, "subject-generic-link-current", "/chat/origin/settings/account")

		getResp := doPendingOIDCRequestWithHeaders(
			t,
			client,
			http.MethodGet,
			ts.URL+"/auth/pending-oidc/"+url.PathEscape(token),
			nil,
			map[string]string{"Authorization": "Bearer " + bearerToken},
		)
		defer getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("GET generic pending status = %d, want 200: %s", getResp.StatusCode, readBody(t, getResp))
		}
		var metadata map[string]any
		if err := json.NewDecoder(getResp.Body).Decode(&metadata); err != nil {
			t.Fatalf("decode generic pending metadata: %v", err)
		}
		if metadata["canLinkCurrent"] != false {
			t.Fatalf("generic canLinkCurrent = %#v, want false", metadata["canLinkCurrent"])
		}

		resp := doPendingOIDCRequestWithHeaders(
			t,
			client,
			http.MethodPost,
			ts.URL+"/auth/pending-oidc/"+url.PathEscape(token)+"/link-current",
			nil,
			map[string]string{"Authorization": "Bearer " + bearerToken},
		)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("POST generic link-current status = %d, want 403: %s", resp.StatusCode, readBody(t, resp))
		}
		linked, err := chattoCore.GetUserByExternalIdentity(t.Context(), "https://issuer.example", "subject-generic-link-current")
		if err != nil {
			t.Fatalf("GetUserByExternalIdentity: %v", err)
		}
		if linked != nil {
			t.Fatalf("generic pending token linked user = %v, want nil", linked)
		}
	})

	t.Run("link current rejects pending token bound to another user", func(t *testing.T) {
		owner, err := chattoCore.CreateUser(t.Context(), core.SystemActorID, "bound-link-owner", "Bound Link Owner", "password123")
		if err != nil {
			t.Fatalf("CreateUser owner: %v", err)
		}
		victim, err := chattoCore.CreateUser(t.Context(), core.SystemActorID, "bound-link-victim", "Bound Link Victim", "password123")
		if err != nil {
			t.Fatalf("CreateUser victim: %v", err)
		}
		victimBearer, err := chattoCore.CreateAuthTokenWithSource(t.Context(), victim.Id, "test")
		if err != nil {
			t.Fatalf("CreateAuthTokenWithSource: %v", err)
		}
		token := createPendingForLinkUser(t, "subject-bound-link-other", "/chat/origin/settings/account", owner.Id)

		getResp := doPendingOIDCRequestWithHeaders(
			t,
			client,
			http.MethodGet,
			ts.URL+"/auth/pending-oidc/"+url.PathEscape(token),
			nil,
			map[string]string{"Authorization": "Bearer " + victimBearer},
		)
		defer getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("GET other-bound pending status = %d, want 200: %s", getResp.StatusCode, readBody(t, getResp))
		}
		var metadata map[string]any
		if err := json.NewDecoder(getResp.Body).Decode(&metadata); err != nil {
			t.Fatalf("decode other-bound pending metadata: %v", err)
		}
		if metadata["canLinkCurrent"] != false {
			t.Fatalf("other-bound canLinkCurrent = %#v, want false", metadata["canLinkCurrent"])
		}

		resp := doPendingOIDCRequestWithHeaders(
			t,
			client,
			http.MethodPost,
			ts.URL+"/auth/pending-oidc/"+url.PathEscape(token)+"/link-current",
			nil,
			map[string]string{"Authorization": "Bearer " + victimBearer},
		)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("POST other-bound link-current status = %d, want 403: %s", resp.StatusCode, readBody(t, resp))
		}
		linked, err := chattoCore.GetUserByExternalIdentity(t.Context(), "https://issuer.example", "subject-bound-link-other")
		if err != nil {
			t.Fatalf("GetUserByExternalIdentity: %v", err)
		}
		if linked != nil {
			t.Fatalf("other-bound pending token linked user = %v, want nil", linked)
		}
	})
}

func doPendingOIDCRequest(t *testing.T, client *http.Client, method, target string, body []byte) *http.Response {
	t.Helper()
	return doPendingOIDCRequestWithHeaders(t, client, method, target, body, nil)
}

func doPendingOIDCRequestWithHeaders(t *testing.T, client *http.Client, method, target string, body []byte, headers map[string]string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	return resp
}

func decodeRedirectURL(t *testing.T, resp *http.Response) string {
	t.Helper()
	var body struct {
		RedirectURL string `json:"redirectUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode redirect response: %v", err)
	}
	return body.RedirectURL
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(data)
}
