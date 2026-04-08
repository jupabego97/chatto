package email

import (
	"errors"
	"testing"

	"hmans.de/chatto/internal/config"
)

func TestMailer_Send_Disabled(t *testing.T) {
	cfg := config.SMTPConfig{
		Enabled: false,
	}
	mailer := NewMailer(cfg)

	err := mailer.Send(Message{
		To:      "test@example.com",
		Subject: "Test",
		Body:    "Test body",
	})

	if !errors.Is(err, ErrSMTPDisabled) {
		t.Errorf("expected ErrSMTPDisabled, got %v", err)
	}
}

func TestMailer_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.SMTPConfig{Enabled: tt.enabled}
			mailer := NewMailer(cfg)
			if got := mailer.IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}
