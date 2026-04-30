package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"hmans.de/chatto/internal/config"
	"hmans.de/chatto/internal/core"
	"hmans.de/chatto/pkg/natsauth"
)

var statsConfigFile string

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Inspect and manage instance stats counters",
	Long: `Commands for working with cached instance counters (e.g., number of
spaces, number of verified users). Counters are read on every gated operation,
so drift can cause limits to mis-fire. Use 'recompute' if you suspect drift.`,
}

var statsRecomputeCmd = &cobra.Command{
	Use:   "recompute",
	Short: "Rebuild instance stats counters from authoritative state",
	Long: `Scans the INSTANCE KV bucket and rewrites the cached counters so they
match reality. Cheap to run; safe at any time. Use after a backup restore or
if you suspect a counter has drifted (e.g., a delete path missed a decrement).

This is the same operation that runs automatically on first startup after
upgrading to a version that supports stats counters.`,
	Run: runStatsRecompute,
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.AddCommand(statsRecomputeCmd)

	statsRecomputeCmd.Flags().StringVarP(&statsConfigFile, "config", "c", "", "path to configuration file (default: chatto.toml)")
}

func runStatsRecompute(cmd *cobra.Command, args []string) {
	cfg, err := config.ReadConfig(statsConfigFile)
	if err != nil {
		log.Fatal("Failed to read configuration", "error", err)
	}

	authOpts, err := natsauth.ConnectOptions(cfg.NATS.Client.NATSAuthConfig())
	if err != nil {
		log.Fatal("Failed to get NATS auth options", "error", err)
	}

	nc, err := nats.Connect(cfg.NATS.Client.URL, authOpts...)
	if err != nil {
		log.Fatal("Failed to connect to NATS", "error", err)
	}
	defer nc.Close()

	ctx := context.Background()

	// Forward CoreConfig values that NewChattoCore expects to be filled in.
	cfg.Core.AuthTokenTTL = cfg.Auth.TokenTTLOrDefault()
	cfg.Core.Replicas = cfg.NATS.ReplicasOrDefault()
	cfg.Core.Limits = cfg.Limits

	chattoCore, err := core.NewChattoCore(ctx, nc, cfg.Core)
	if err != nil {
		log.Fatal("Failed to initialize core", "error", err)
	}

	log.Info("Recomputing instance stats from authoritative state...")
	if err := chattoCore.RecomputeStats(ctx); err != nil {
		log.Fatal("Recompute failed", "error", err)
	}

	// Read back and report.
	spaces, _ := chattoCore.GetStat(ctx, core.StatSpaces)
	users, _ := chattoCore.GetStat(ctx, core.StatVerifiedUsers)
	fmt.Printf("\nRecomputed counters:\n")
	fmt.Printf("  spaces:         %d\n", spaces)
	fmt.Printf("  verified_users: %d\n", users)
}
