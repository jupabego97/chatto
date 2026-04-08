package embedded_nats

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"hmans.de/chatto/internal/config"
)

// Start creates, starts, and registers the embedded NATS server with the errgroup.
// It blocks until the server is ready for connections, then returns.
// Shutdown is handled by the errgroup goroutine when ctx is cancelled.
func Start(ctx context.Context, g *errgroup.Group, cfg *config.EmbeddedNATSConfig) (*server.Server, error) {
	logger := log.WithPrefix("server.NATS")

	ns, err := createServer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	ns.Start()

	// Wait for server to be ready for connections
	if !ns.ReadyForConnections(4 * time.Second) {
		return nil, fmt.Errorf("server failed to start within timeout")
	}

	if cfg.Port == 0 {
		logger.Info("Embedded NATS server is ready (in-process only, no TCP listener)")
	} else {
		logger.Info("Embedded NATS server is ready",
			"address", fmt.Sprintf("%s:%d", cfg.BindAddressOrDefault(), cfg.Port),
			"auth", cfg.AuthToken != "")
	}

	// Register shutdown handler with errgroup
	g.Go(func() error {
		<-ctx.Done()
		logger.Info("Shutting down embedded NATS server")
		ns.Shutdown()
		ns.WaitForShutdown()
		logger.Info("Embedded NATS server shutdown complete")
		return nil
	})

	return ns, nil
}

// createServer creates an embedded NATS server configured from chatto.toml.
// Use InProcessConnectOption for secure in-process connections.
// When Port > 0, a TCP listener is enabled with token authentication.
func createServer(cfg *config.EmbeddedNATSConfig) (*server.Server, error) {
	opts := &server.Options{
		JetStream: true,
		StoreDir:  cfg.DataDir,
		NoSigs:    true, // Let the app handle signals
	}

	// TCP client port configuration
	if cfg.Port == 0 {
		opts.DontListen = true
	} else {
		opts.Port = cfg.Port
		opts.Host = cfg.BindAddressOrDefault()
		// Enable token auth when configured
		if cfg.AuthToken != "" {
			opts.Authorization = cfg.AuthToken
		}
	}

	// HTTP monitoring port configuration
	if cfg.HTTPPort > 0 {
		opts.HTTPPort = cfg.HTTPPort
		opts.HTTPHost = cfg.BindAddressOrDefault()
	}

	return server.NewServer(opts)
}

// InProcessConnectOption returns a NATS connection option that connects
// directly to the embedded server without going through TCP.
func InProcessConnectOption(srv *server.Server) nats.Option {
	return nats.InProcessServer(srv)
}
