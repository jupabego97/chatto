package rbac

import (
	"errors"
	"testing"
)

func TestValidateSpaceRoleName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		// Valid space role names - lowercase letters only
		{"simple name", "admin", nil},
		{"moderator", "moderator", nil},
		{"single char", "a", nil},
		{"max length", "abcdefghijklmnopqrstuvwxyzabcdef", nil}, // 32 chars

		// Invalid - reserved name
		{"reserved instance", "instance", ErrReservedRoleName},

		// Invalid - contains numbers
		{"with numbers", "tier2", ErrInvalidRoleName},
		{"ends with number", "admin1", ErrInvalidRoleName},

		// Invalid - contains dashes
		{"with dashes", "power-user", ErrInvalidRoleName},
		{"hyphenated", "super-admin", ErrInvalidRoleName},

		// Invalid - other characters
		{"empty", "", ErrInvalidRoleName},
		{"starts with number", "2tier", ErrInvalidRoleName},
		{"starts with dash", "-admin", ErrInvalidRoleName},
		{"uppercase", "Admin", ErrInvalidRoleName},
		{"mixed case", "PowerUser", ErrInvalidRoleName},
		{"spaces", "power user", ErrInvalidRoleName},
		{"underscore", "power_user", ErrInvalidRoleName},
		{"dot", "power.user", ErrInvalidRoleName},
		{"too long", "abcdefghijklmnopqrstuvwxyzabcdefg", ErrInvalidRoleName}, // 33 chars
		{"special chars", "admin!", ErrInvalidRoleName},
		{"unicode", "admin\u00e9", ErrInvalidRoleName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpaceRoleName(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateSpaceRoleName(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateSpaceRoleName(%q) expected error %v, got nil", tt.input, tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateSpaceRoleName(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateInstanceRoleName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		// Valid instance role names
		{"instance admin", "instance-admin", nil},
		{"instance editor", "instance-editor", nil},
		{"instance custom", "instance-customrole", nil},
		{"instance single char", "instance-a", nil},
		{"instance max suffix", "instance-abcdefghijklmnopqrstuvw", nil}, // 23 char suffix = 32 total

		// Invalid - doesn't start with instance-
		{"no prefix", "admin", ErrInvalidInstanceRoleName},
		{"wrong prefix", "inst-admin", ErrInvalidInstanceRoleName},

		// Invalid - suffix contains numbers
		{"suffix with number", "instance-admin2", ErrInvalidInstanceRoleName},
		{"suffix with numbers", "instance-tier2support", ErrInvalidInstanceRoleName},

		// Invalid - suffix contains dashes
		{"suffix with dash", "instance-super-admin", ErrInvalidInstanceRoleName},

		// Invalid - other issues
		{"empty suffix", "instance-", ErrInvalidInstanceRoleName},
		{"uppercase suffix", "instance-Admin", ErrInvalidInstanceRoleName},
		{"suffix too long", "instance-abcdefghijklmnopqrstuvwx", ErrInvalidInstanceRoleName}, // 24 char suffix = 33 total
		{"just instance", "instance", ErrInvalidInstanceRoleName},
		{"empty", "", ErrInvalidInstanceRoleName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstanceRoleName(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateInstanceRoleName(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateInstanceRoleName(%q) expected error %v, got nil", tt.input, tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateInstanceRoleName(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateRoleName(t *testing.T) {
	// ValidateRoleName auto-detects role type based on prefix

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Space roles (valid)
		{"space admin", "admin", false},
		{"space moderator", "moderator", false},

		// Space roles (invalid)
		{"space with numbers", "tier2", true},
		{"space with dash", "power-user", true},
		{"space reserved", "instance", true},

		// Instance roles (valid)
		{"instance admin", "instance-admin", false},
		{"instance editor", "instance-editor", false},

		// Instance roles (invalid)
		{"instance with number", "instance-admin2", true},
		{"instance with dash", "instance-super-admin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoleName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRoleName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
