package crypto

import (
	"agropal/config"
	"agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/identity"
	"agropal/pkg"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func newTestJWTProvider(jwtSecret, gcmSecret string) *JWTProvider {
	cfg := &config.Config{}
	cfg.Secret.Jwt = jwtSecret
	cfg.Secret.GcmToken = gcmSecret
	logger := pkg.NewLogger(*cfg)

	return &JWTProvider{
		logger:    logger,
		config:    cfg,
		enc:       NewEncryption(logger, cfg),
		generator: identity.NewGenerator(),
	}
}

func TestGenerateTokenPair(t *testing.T) {
	ctx := context.Background()
	claimInfo := shared.ClaimInfo{
		UserID: uuid.New(),
		Email:  "user@example.com",
		Role:   "farmer",
	}

	t.Run("Success", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		pair, err := j.GenerateTokenPair(ctx, claimInfo)
		if err != nil {
			t.Fatalf("GenerateTokenPair() error = %v", err)
		}

		if pair.AccessTokenStr == "" {
			t.Error("GenerateTokenPair() access token is empty")
		}
		if pair.RefreshToken.Token == "" {
			t.Error("GenerateTokenPair() refresh token is empty")
		}
	})

	t.Run("Access Token Is Valid And Contains Claims", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		pair, err := j.GenerateTokenPair(ctx, claimInfo)
		if err != nil {
			t.Fatalf("GenerateTokenPair() error = %v", err)
		}

		got, err := j.ValidateToken(ctx, shared.TypeAccess, pair.AccessTokenStr)
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}

		if got.UserID != claimInfo.UserID || got.Email != claimInfo.Email || got.Role != claimInfo.Role {
			t.Errorf("ValidateToken() = %+v, want %+v", got, claimInfo)
		}
	})

	t.Run("Refresh Token Is Encrypted And Round Trips", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		pair, err := j.GenerateTokenPair(ctx, claimInfo)
		if err != nil {
			t.Fatalf("GenerateTokenPair() error = %v", err)
		}

		got, err := j.ValidateToken(ctx, shared.TypeRefresh, pair.RefreshToken.Token)
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}

		if got.UserID != claimInfo.UserID {
			t.Errorf("ValidateToken() UserID = %v, want %v", got.UserID, claimInfo.UserID)
		}
	})

	t.Run("Refresh Token Metadata Populated From Context", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		metaCtx := shared.WithMetadata(ctx, shared.KeyUserAgent, "test-agent")
		metaCtx = shared.WithMetadata(metaCtx, shared.KeyIPAddress, "127.0.0.1")

		pair, err := j.GenerateTokenPair(metaCtx, claimInfo)
		if err != nil {
			t.Fatalf("GenerateTokenPair() error = %v", err)
		}

		if pair.RefreshToken.UserAgent != "test-agent" {
			t.Errorf("RefreshToken.UserAgent = %q, want %q", pair.RefreshToken.UserAgent, "test-agent")
		}
		if pair.RefreshToken.IPAddress != "127.0.0.1" {
			t.Errorf("RefreshToken.IPAddress = %q, want %q", pair.RefreshToken.IPAddress, "127.0.0.1")
		}
		if pair.RefreshToken.UserID != claimInfo.UserID {
			t.Errorf("RefreshToken.UserID = %v, want %v", pair.RefreshToken.UserID, claimInfo.UserID)
		}
		if pair.RefreshToken.ID == "" {
			t.Error("RefreshToken.ID is empty")
		}
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	claimInfo := shared.ClaimInfo{
		UserID: uuid.New(),
		Email:  "user@example.com",
		Role:   "farmer",
	}

	t.Run("Invalid Token String", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		_, err := j.ValidateToken(ctx, shared.TypeAccess, "not-a-jwt")
		if err == nil {
			t.Error("ValidateToken() error = nil, want error for malformed token")
		}
	})

	t.Run("Wrong Signing Secret", func(t *testing.T) {
		signed, err := newTestJWTProvider("secret-one", "gcm-secret").signToken(claimInfo, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("signToken() error = %v", err)
		}

		_, err = newTestJWTProvider("secret-two", "gcm-secret").ValidateToken(ctx, shared.TypeAccess, signed)
		if err == nil {
			t.Error("ValidateToken() error = nil, want error when validating with the wrong secret")
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		signed, err := j.signToken(claimInfo, time.Now().Add(-time.Hour))
		if err != nil {
			t.Fatalf("signToken() error = %v", err)
		}

		_, err = j.ValidateToken(ctx, shared.TypeAccess, signed)
		if err == nil {
			t.Error("ValidateToken() error = nil, want error for expired token")
		}
	})

	t.Run("Unexpected Signing Method", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		claims := jwtClaims{
			ClaimInfo: claimInfo,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		signed, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("SignedString() error = %v", err)
		}

		_, err = j.ValidateToken(ctx, shared.TypeAccess, signed)
		if err == nil {
			t.Error("ValidateToken() error = nil, want error for unexpected signing method")
		}
	})

	t.Run("Refresh Token Fails When Not Encrypted", func(t *testing.T) {
		j := newTestJWTProvider("jwt-secret", "gcm-secret")

		signed, err := j.signToken(claimInfo, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("signToken() error = %v", err)
		}

		_, err = j.ValidateToken(ctx, shared.TypeRefresh, signed)
		if !errors.Is(err, auth.ErrTokenInvalid) {
			t.Errorf("ValidateToken() error = %v, want %v", err, auth.ErrTokenInvalid)
		}
	})
}
