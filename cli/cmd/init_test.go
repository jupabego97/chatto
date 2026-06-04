package cmd

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"hmans.de/chatto/internal/config"
)

func TestInitGeneratesCoreSecret(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	originalConfigFile := initConfigFile
	initConfigFile = ""
	t.Cleanup(func() { initConfigFile = originalConfigFile })

	initCmd.Run(initCmd, nil)

	cfg, err := config.ReadConfig(filepath.Join(tmpDir, "chatto.toml"))
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if len(cfg.Core.SecretKey) != 64 {
		t.Fatalf("generated core secret length = %d, want 64", len(cfg.Core.SecretKey))
	}
	if _, err := hex.DecodeString(cfg.Core.SecretKey); err != nil {
		t.Fatalf("generated core secret should be hex: %v", err)
	}
	if cfg.Core.IdentityClaims.ActiveKeyID != "v1" {
		t.Fatalf("generated identity claim active key = %q, want v1", cfg.Core.IdentityClaims.ActiveKeyID)
	}
	if len(cfg.Core.IdentityClaims.Keys) != 1 {
		t.Fatalf("generated identity claim keys = %d, want 1", len(cfg.Core.IdentityClaims.Keys))
	}
	key := cfg.Core.IdentityClaims.Keys[0]
	if key.ID != "v1" {
		t.Fatalf("generated identity claim key ID = %q, want v1", key.ID)
	}
	if len(key.Secret) != 64 {
		t.Fatalf("generated identity claim secret length = %d, want 64", len(key.Secret))
	}
	if _, err := hex.DecodeString(key.Secret); err != nil {
		t.Fatalf("generated identity claim secret should be hex: %v", err)
	}
}
