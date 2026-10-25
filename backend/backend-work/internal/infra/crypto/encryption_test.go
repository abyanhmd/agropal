package crypto

import (
	"agropal/config"
	"agropal/pkg"
	"context"
	"encoding/base64"
	"errors"
	"testing"
)

func newTestEncryption(secret string) *Encryption {
	cfg := &config.Config{}
	cfg.Secret.GcmToken = secret
	logger := pkg.NewLogger(*cfg)
	return NewEncryption(logger, cfg)
}

func TestEncryptAES(t *testing.T) {
	ctx := context.Background()

	t.Run("Round Trip", func(t *testing.T) {
		e := newTestEncryption("test-secret")
		plainText := "super-secret-refresh-token"

		cipherText, err := e.EncryptAES(ctx, EncRefreshToken, plainText)
		if err != nil {
			t.Fatalf("EncryptAES() error = %v", err)
		}

		got, err := e.DecryptAES(ctx, EncRefreshToken, cipherText)
		if err != nil {
			t.Fatalf("DecryptAES() error = %v", err)
		}

		if got != plainText {
			t.Errorf("DecryptAES() = %q, want %q", got, plainText)
		}
	})

	t.Run("Empty Plain Text", func(t *testing.T) {
		e := newTestEncryption("test-secret")

		cipherText, err := e.EncryptAES(ctx, EncRefreshToken, "")
		if err != nil {
			t.Fatalf("EncryptAES() error = %v", err)
		}

		got, err := e.DecryptAES(ctx, EncRefreshToken, cipherText)
		if err != nil {
			t.Fatalf("DecryptAES() error = %v", err)
		}

		if got != "" {
			t.Errorf("DecryptAES() = %q, want empty string", got)
		}
	})

	t.Run("Non Deterministic Output", func(t *testing.T) {
		e := newTestEncryption("test-secret")
		plainText := "same-input-twice"

		first, err := e.EncryptAES(ctx, EncRefreshToken, plainText)
		if err != nil {
			t.Fatalf("EncryptAES() error = %v", err)
		}
		second, err := e.EncryptAES(ctx, EncRefreshToken, plainText)
		if err != nil {
			t.Fatalf("EncryptAES() error = %v", err)
		}

		if first == second {
			t.Errorf("EncryptAES() produced identical ciphertext for two calls, want distinct nonces to yield distinct output")
		}
	})
}

func TestDecryptAES(t *testing.T) {
	ctx := context.Background()

	t.Run("Wrong Key", func(t *testing.T) {
		cipherText, err := newTestEncryption("key-one").EncryptAES(ctx, EncRefreshToken, "some-token")
		if err != nil {
			t.Fatalf("EncryptAES() error = %v", err)
		}

		_, err = newTestEncryption("key-two").DecryptAES(ctx, EncRefreshToken, cipherText)
		if err == nil {
			t.Error("DecryptAES() error = nil, want error when decrypting with the wrong key")
		}
	})

	t.Run("Tampered Cipher Text", func(t *testing.T) {
		e := newTestEncryption("test-secret")

		cipherText, err := e.EncryptAES(ctx, EncRefreshToken, "some-token")
		if err != nil {
			t.Fatalf("EncryptAES() error = %v", err)
		}

		raw, err := base64.StdEncoding.DecodeString(cipherText)
		if err != nil {
			t.Fatalf("failed to decode ciphertext: %v", err)
		}
		raw[len(raw)-1] ^= 0xFF
		tampered := base64.StdEncoding.EncodeToString(raw)

		_, err = e.DecryptAES(ctx, EncRefreshToken, tampered)
		if err == nil {
			t.Error("DecryptAES() error = nil, want error for tampered ciphertext (GCM auth tag should fail)")
		}
	})

	t.Run("Invalid Base64", func(t *testing.T) {
		e := newTestEncryption("test-secret")

		_, err := e.DecryptAES(ctx, EncRefreshToken, "not-valid-base64!!!")
		if err == nil {
			t.Error("DecryptAES() error = nil, want error for invalid base64 input")
		}
	})

	t.Run("Too Short To Contain Nonce", func(t *testing.T) {
		e := newTestEncryption("test-secret")

		short := base64.StdEncoding.EncodeToString([]byte("short"))

		_, err := e.DecryptAES(ctx, EncRefreshToken, short)
		if !errors.Is(err, ErrCipherTextTooShort) {
			t.Errorf("DecryptAES() error = %v, want %v", err, ErrCipherTextTooShort)
		}
	})
}
