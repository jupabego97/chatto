// Package natsauth provides authentication option builders for NATS connections.
// It supports token, username/password, credentials file, and NKey authentication.
package natsauth

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

// AuthMethod defines how to authenticate with NATS.
type AuthMethod string

const (
	AuthNone        AuthMethod = "none"        // No authentication
	AuthToken       AuthMethod = "token"       // Simple bearer token
	AuthUserPass    AuthMethod = "userpass"     // Username/password
	AuthCredentials AuthMethod = "credentials" // JWT credentials file
	AuthNKey        AuthMethod = "nkey"         // NKey seed
)

// Config holds the parameters needed to build NATS authentication options.
type Config struct {
	AuthMethod      AuthMethod
	Token           string
	Username        string
	Password        string
	CredentialsFile string
	NKeySeed        string
}

// ConnectOptions returns NATS connection options for the given auth configuration.
func ConnectOptions(cfg Config) ([]nats.Option, error) {
	switch cfg.AuthMethod {
	case AuthNone, "":
		return nil, nil

	case AuthToken:
		if cfg.Token == "" {
			return nil, fmt.Errorf("nats auth: token is required for token method")
		}
		return []nats.Option{nats.Token(cfg.Token)}, nil

	case AuthUserPass:
		if cfg.Username == "" {
			return nil, fmt.Errorf("nats auth: username is required for userpass method")
		}
		return []nats.Option{nats.UserInfo(cfg.Username, cfg.Password)}, nil

	case AuthCredentials:
		if cfg.CredentialsFile == "" {
			return nil, fmt.Errorf("nats auth: credentials_file is required for credentials method")
		}
		return []nats.Option{nats.UserCredentials(cfg.CredentialsFile)}, nil

	case AuthNKey:
		if cfg.NKeySeed == "" {
			return nil, fmt.Errorf("nats auth: nkey_seed is required for nkey method")
		}
		opt, err := nkeyOption(cfg.NKeySeed)
		if err != nil {
			return nil, err
		}
		return []nats.Option{opt}, nil

	default:
		return nil, fmt.Errorf("nats auth: unknown method %q", cfg.AuthMethod)
	}
}

// nkeyOption creates a NATS option for NKey authentication.
func nkeyOption(seed string) (nats.Option, error) {
	kp, err := nkeys.FromSeed([]byte(seed))
	if err != nil {
		return nil, fmt.Errorf("nats auth: invalid nkey seed: %w", err)
	}

	pubKey, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("nats auth: failed to get public key: %w", err)
	}

	return nats.Nkey(pubKey, func(nonce []byte) ([]byte, error) {
		return kp.Sign(nonce)
	}), nil
}

// PublicKeyFromSeed extracts the public key from an NKey seed.
// This is useful for generating the nats-server.conf authorization section.
func PublicKeyFromSeed(seed string) (string, error) {
	kp, err := nkeys.FromSeed([]byte(seed))
	if err != nil {
		return "", fmt.Errorf("invalid nkey seed: %w", err)
	}
	return kp.PublicKey()
}

// GenerateUserNKey generates a new user NKey pair.
// Returns the seed (private, for config) and public key (for nats-server.conf).
func GenerateUserNKey() (seed, publicKey string, err error) {
	kp, err := nkeys.CreateUser()
	if err != nil {
		return "", "", fmt.Errorf("failed to create user nkey: %w", err)
	}

	seedBytes, err := kp.Seed()
	if err != nil {
		return "", "", fmt.Errorf("failed to get seed: %w", err)
	}

	pubKey, err := kp.PublicKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to get public key: %w", err)
	}

	return string(seedBytes), pubKey, nil
}
