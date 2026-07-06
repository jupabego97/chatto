package http_server

import (
	"context"
	"testing"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

func TestNormalizeLoopbackHost(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"http://localhost:5173", "http://127.0.0.1:5173"},
		{"http://localhost", "http://127.0.0.1"},
		{"http://localhost:5173/auth/atproto/callback", "http://127.0.0.1:5173/auth/atproto/callback"},
		{"http://127.0.0.1:5173", "http://127.0.0.1:5173"},
		{"https://chatto.example.com", "https://chatto.example.com"},
		{"https://chatto.example.com/auth/atproto/callback", "https://chatto.example.com/auth/atproto/callback"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := normalizeLoopbackHost(tc.in); got != tc.want {
				t.Errorf("normalizeLoopbackHost(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsLocalhostURL(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"http://localhost:8080", true},
		{"https://localhost", true},
		{"http://127.0.0.1:3000", true},
		{"http://[::1]:8080", true},
		{"https://chatto.example.com", false},
		{"https://localhost.example.com", false},
		{"", false},
		{"not a url", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := isLocalhostURL(tc.in); got != tc.want {
				t.Errorf("isLocalhostURL(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestDeleteLocalATProtoSessionOnlyDeletesStoreEntry(t *testing.T) {
	did, err := syntax.ParseDID("did:plc:alice")
	if err != nil {
		t.Fatalf("ParseDID: %v", err)
	}
	store := &recordingATProtoStore{}

	if err := deleteLocalATProtoSession(context.Background(), store, &oauth.ClientSessionData{
		AccountDID: did,
		SessionID:  "session-123",
	}); err != nil {
		t.Fatalf("deleteLocalATProtoSession: %v", err)
	}

	if store.deletedDID != "did:plc:alice" || store.deletedSessionID != "session-123" {
		t.Fatalf("deleted = %q/%q, want did:plc:alice/session-123", store.deletedDID, store.deletedSessionID)
	}
}

type recordingATProtoStore struct {
	deletedDID       string
	deletedSessionID string
}

func (s *recordingATProtoStore) GetSession(context.Context, syntax.DID, string) (*oauth.ClientSessionData, error) {
	return nil, nil
}

func (s *recordingATProtoStore) SaveSession(context.Context, oauth.ClientSessionData) error {
	return nil
}

func (s *recordingATProtoStore) DeleteSession(_ context.Context, did syntax.DID, sessionID string) error {
	s.deletedDID = did.String()
	s.deletedSessionID = sessionID
	return nil
}

func (s *recordingATProtoStore) GetAuthRequestInfo(context.Context, string) (*oauth.AuthRequestData, error) {
	return nil, nil
}

func (s *recordingATProtoStore) SaveAuthRequestInfo(context.Context, oauth.AuthRequestData) error {
	return nil
}

func (s *recordingATProtoStore) DeleteAuthRequestInfo(context.Context, string) error {
	return nil
}
