//go:build !bootstrap

package http_server

// devATProtoURLOverride is empty in release builds, so ATProto OAuth always
// uses the configured webserver.url. Release deployments must be publicly
// reachable for the OAuth flow to work; there is no escape hatch.
const devATProtoURLOverride = ""
