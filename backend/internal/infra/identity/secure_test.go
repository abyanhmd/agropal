package identity

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestGenerateHexID(t *testing.T) {
	g := NewGenerator()

	t.Run("Produces Valid Hex String Of Expected Length", func(t *testing.T) {
		id, err := g.GenerateHexID()
		if err != nil {
			t.Fatalf("GenerateHexID() error = %v", err)
		}

		if len(id) != 32 {
			t.Errorf("GenerateHexID() length = %d, want %d", len(id), 32)
		}

		if _, err := hex.DecodeString(id); err != nil {
			t.Errorf("GenerateHexID() = %q, not valid hex: %v", id, err)
		}
	})

	t.Run("Produces Distinct Values Across Calls", func(t *testing.T) {
		first, err := g.GenerateHexID()
		if err != nil {
			t.Fatalf("GenerateHexID() error = %v", err)
		}
		second, err := g.GenerateHexID()
		if err != nil {
			t.Fatalf("GenerateHexID() error = %v", err)
		}

		if first == second {
			t.Errorf("GenerateHexID() produced identical values %q on consecutive calls, want distinct", first)
		}
	})
}

func TestGenerateSecureID(t *testing.T) {
	g := NewGenerator()

	t.Run("Produces Valid Base64 URL Encoded String Of Expected Length", func(t *testing.T) {
		id, err := g.GenerateSecureID()
		if err != nil {
			t.Fatalf("GenerateSecureID() error = %v", err)
		}

		if len(id) != 30 {
			t.Errorf("GenerateSecureID() length = %d, want %d", len(id), 30)
		}

		if _, err := base64.RawURLEncoding.DecodeString(id); err != nil {
			t.Errorf("GenerateSecureID() = %q, not valid base64 URL encoding: %v", id, err)
		}
	})

	t.Run("Produces Distinct Values Across Calls", func(t *testing.T) {
		first, err := g.GenerateSecureID()
		if err != nil {
			t.Fatalf("GenerateSecureID() error = %v", err)
		}
		second, err := g.GenerateSecureID()
		if err != nil {
			t.Fatalf("GenerateSecureID() error = %v", err)
		}

		if first == second {
			t.Errorf("GenerateSecureID() produced identical values %q on consecutive calls, want distinct", first)
		}
	})
}
