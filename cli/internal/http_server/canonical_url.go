package http_server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// canonicalServerOrigin returns the configured public origin (scheme + host)
// for a Chatto server. Empty or malformed config means canonicalization is off.
func canonicalServerOrigin(webserverURL string) string {
	if webserverURL == "" {
		return ""
	}

	parsed, err := url.Parse(webserverURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}

	return scheme + "://" + strings.ToLower(parsed.Host)
}

func incomingRequestOrigin(c *gin.Context) string {
	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = firstForwardedHeaderValue(proto)
	} else if c.Request.TLS != nil {
		scheme = "https"
	}

	host := c.Request.Host
	if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
		host = firstForwardedHeaderValue(forwardedHost)
	}

	return scheme + "://" + strings.ToLower(strings.TrimSpace(host))
}

func firstForwardedHeaderValue(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(value, ",")[0]))
}

func isHealthProbePath(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func (s *HTTPServer) canonicalRedirectMiddleware() gin.HandlerFunc {
	canonicalOrigin := canonicalServerOrigin(s.config.Webserver.URL)
	return func(c *gin.Context) {
		if canonicalOrigin == "" ||
			isHealthProbePath(c.Request.URL.Path) ||
			incomingRequestOrigin(c) == canonicalOrigin {
			c.Next()
			return
		}

		target := canonicalOrigin + c.Request.URL.RequestURI()
		c.Redirect(http.StatusPermanentRedirect, target)
		c.Abort()
	}
}
