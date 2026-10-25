package config

import (
	"strings"
	"testing"
)

func TestValidateSecrets(t *testing.T) {
	long := strings.Repeat("a", minSecretLength)

	tests := []struct {
		name    string
		jwt     string
		gcm     string
		wantErr bool
	}{
		{"both long enough", long, long, false},
		{"empty jwt", "", long, true},
		{"short gcm", long, "short", true},
		{"both missing", "", "", true},
		{"placeholder jwt", "change-me-run-openssl-rand-hex-32", long, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			cfg.Secret.Jwt = tt.jwt
			cfg.Secret.GcmToken = tt.gcm

			err := cfg.ValidateSecrets()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
