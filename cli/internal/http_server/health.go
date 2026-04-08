package http_server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupHealthRoutes registers health check endpoints for Kubernetes probes.
func (s *HTTPServer) setupHealthRoutes() {
	// Liveness probe - is the server process alive?
	// Returns 200 if the HTTP server is running.
	s.router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Readiness probe - is the server ready to accept traffic?
	// Checks NATS connectivity and JetStream initialization.
	s.router.GET("/readyz", func(c *gin.Context) {
		// Check NATS connection
		if s.nc == nil || !s.nc.IsConnected() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "NATS not connected",
			})
			return
		}

		// Check JetStream resources are initialized
		if s.core != nil {
			if err := s.core.Ready(c.Request.Context()); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not ready",
					"reason": err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
