package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/mzwrt/dujiao-next/internal/config"
)

func TestValidatePassword_PCIDSSMinLength(t *testing.T) {
	// PCI-DSS 8.2.3 requires minimum 7 characters even with empty policy.
	zeroCfg := config.PasswordPolicyConfig{}

	tests := []struct {
		name    string
		pass    string
		wantErr bool
	}{
		{"too_short_4", "abcd", true},
		{"too_short_6", "abcdef", true},
		{"exactly_7", "abcdefg", false},
		{"exactly_8", "abcdefgh", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(zeroCfg, tt.pass)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q but got nil", tt.pass)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.pass, err)
			}
			if tt.wantErr && err != nil {
				if !errors.Is(err, ErrWeakPassword) {
					t.Fatalf("expected ErrWeakPassword, got %v", err)
				}
			}
		})
	}
}

func TestValidatePassword_ConfigHigherThanPCI(t *testing.T) {
	// When config sets MinLength > 7, that takes precedence.
	cfg := config.PasswordPolicyConfig{MinLength: 10}
	err := validatePassword(cfg, "abcdefgh") // 8 chars < 10
	if err == nil {
		t.Fatal("expected error for 8-char password with MinLength=10")
	}
	err = validatePassword(cfg, "abcdefghij") // 10 chars
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePassword_BcryptMax(t *testing.T) {
	long := strings.Repeat("a", BcryptMaxPasswordBytes+1)
	err := validatePassword(config.PasswordPolicyConfig{}, long)
	if err == nil {
		t.Fatal("expected error for password exceeding bcrypt max")
	}
}
