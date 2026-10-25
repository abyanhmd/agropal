package identity

import (
	"regexp"
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	g := NewGenerator()

	t.Run("Produces Six Digit String", func(t *testing.T) {
		otp := g.GenerateOTP()

		if len(otp) != 6 {
			t.Errorf("GenerateOTP() length = %d, want %d", len(otp), 6)
		}

		if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(otp) {
			t.Errorf("GenerateOTP() = %q, want a 6 digit numeric string", otp)
		}
	})

	t.Run("Produces Distinct Values Across Calls", func(t *testing.T) {
		seen := make(map[string]bool)
		for range 20 {
			seen[g.GenerateOTP()] = true
		}

		if len(seen) < 2 {
			t.Errorf("GenerateOTP() produced only %d distinct value(s) across 20 calls, want variation", len(seen))
		}
	})
}

func TestGenerateReferral(t *testing.T) {
	g := NewGenerator()

	t.Run("Produces Eight Character Alphanumeric String", func(t *testing.T) {
		referral := g.GenerateReferral()

		if len(referral) != 8 {
			t.Errorf("GenerateReferral() length = %d, want %d", len(referral), 8)
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9]{8}$`).MatchString(referral) {
			t.Errorf("GenerateReferral() = %q, want an 8 character alphanumeric string", referral)
		}
	})

	t.Run("Produces Distinct Values Across Calls", func(t *testing.T) {
		first := g.GenerateReferral()
		second := g.GenerateReferral()

		if first == second {
			t.Errorf("GenerateReferral() produced identical values %q on consecutive calls, want distinct", first)
		}
	})
}
