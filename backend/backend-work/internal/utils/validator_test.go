package utils

import (
	dAuth "agropal/internal/domain/auth"
	"errors"
	"testing"
)

func TestValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid simple", "user@example.com", true},
		{"valid with subdomain", "user@mail.example.co.uk", true},
		{"valid with plus tag", "user+tag@example.com", true},
		{"missing @", "userexample.com", false},
		{"missing domain", "user@", false},
		{"missing local part", "@example.com", false},
		{"empty string", "", false},
		{"just whitespace", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidEmail(tt.email); got != tt.want {
				t.Errorf("ValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid strong password", "Str0ng!Pass", nil},
		{"too short", "Sh0rt!", dAuth.ErrPasswordTooShort},
		{"too long", "Aa1!" + stringOfLength(61), dAuth.ErrPasswordTooLong},
		{"no lowercase", "STRONG1!PASS", dAuth.ErrPasswordHasNoLower},
		{"no uppercase", "str0ng!pass", dAuth.ErrPasswordHasNoUpper},
		{"no number", "Strong!Pass", dAuth.ErrPasswordHasNoNumber},
		{"no special char", "Str0ngPassword", dAuth.ErrPasswordHasNoSpecial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidPassword(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidPassword(%q) = %v, want %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// stringOfLength returns a string of lowercase 'a's of the given length,
// used to pad a password past the max-length boundary in tests.
func stringOfLength(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
