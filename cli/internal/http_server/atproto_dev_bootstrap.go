//go:build bootstrap

package http_server

// devATProtoURLOverride forces the ATProto OAuth flow onto the loopback
// client form in dev-image builds. orb.local-style dev hostnames can't be
// reached from a real PDS, and the loopback form (http://localhost?…) skips
// the client-metadata fetch entirely. The redirect URI must be 127.0.0.1
// (or [::1]) per the ATProto spec — `localhost` is rejected — so the dev
// frontend must be reachable at the URL below. compose.yml binds the Vite
// dev server to 127.0.0.1:5173 specifically for this flow.
const devATProtoURLOverride = "http://127.0.0.1:5173"
