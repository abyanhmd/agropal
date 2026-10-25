package config

import (
	"fmt"
	"strings"
)

// minSecretLength is the shortest signing/encryption secret the API will start with.
// `openssl rand -hex 32` gives 64 characters.
const minSecretLength = 32

// ValidateSecrets fails fast on a missing or short JWT/GCM secret, so a misconfigured deployment
// cannot sign tokens with an empty or guessable key. LoadConfig does not call it, because the
// migration and email commands do not use these secrets.
func (c *Config) ValidateSecrets() error {
	secrets := map[string]string{
		"SECRET_JWT":       c.Secret.Jwt,
		"SECRET_GCM_TOKEN": c.Secret.GcmToken,
	}

	var problems []string
	for name, value := range secrets {
		switch {
		case len(value) < minSecretLength:
			problems = append(problems, fmt.Sprintf("%s must be at least %d characters", name, minSecretLength))
		case strings.HasPrefix(value, "change-me"):
			problems = append(problems, fmt.Sprintf("%s is still the placeholder from .env.example", name))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("insecure configuration: %s", strings.Join(problems, "; "))
	}
	return nil
}
