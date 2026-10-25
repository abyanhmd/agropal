package identity

import (
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

func (g *Generator) GenerateHexID() (string, error) {
	// Example: "9a3bf28cc71e4d02b13dfbb48143629f"
	n := 16

	b := make([]byte, n)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (g *Generator) GenerateSecureID() (string, error) {
	// Example: "y0Df28cdSuWmk3hKkmx8zA"
	n := 22

	b := make([]byte, n)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return "", err
	}

	// base64 URL-safe without padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}
