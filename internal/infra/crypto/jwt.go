package crypto

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/identity"
	"agropal/pkg"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	dAuth.ClaimInfo
	jwt.RegisteredClaims
}

type JWTProvider struct {
	logger    *pkg.Logger
	config    *config.Config
	enc       *Encryption
	generator *identity.Generator
}

func NewJWTProvider(
	logger *pkg.Logger,
	config *config.Config,
	enc *Encryption,
	generator *identity.Generator,
) dAuth.TokenProvider {
	return &JWTProvider{
		logger:    logger,
		config:    config,
		enc:       enc,
		generator: generator,
	}
}

func (j *JWTProvider) GenerateTokenPair(ctx context.Context, claimInfo dAuth.ClaimInfo) (dAuth.TokenPair, error) {
	now := time.Now().UTC()

	accessDuration := time.Duration(30)
	accessExpiry := now.Add(accessDuration * time.Minute)

	refreshDuration := time.Duration(7)
	refreshExpiry := now.Add(refreshDuration * 24 * time.Hour)

	signedAccess, err := j.signToken(claimInfo, accessExpiry)
	if err != nil {
		j.logger.Error(ctx, shared.System, err.Error())
		return dAuth.TokenPair{}, fmt.Errorf("failed to sign token. %w", err)
	}

	signedRefresh, err := j.signToken(claimInfo, refreshExpiry)
	if err != nil {
		j.logger.Error(ctx, shared.System, err.Error())
		return dAuth.TokenPair{}, fmt.Errorf("failed to sign token. %w", err)
	}

	hashedRefresh, err := j.enc.EncryptAES(ctx, EncRefreshToken, signedRefresh)
	if err != nil {
		j.logger.Error(ctx, shared.System, err.Error())
		return dAuth.TokenPair{}, fmt.Errorf("failed to encrypt token. %w", err)
	}

	newCookieID, err := j.generator.GenerateHexID()
	if err != nil {
		j.logger.Error(ctx, shared.System, err.Error())
		return dAuth.TokenPair{}, fmt.Errorf("failed to generate cookie id. %w", err)
	}
	refreshToken := dAuth.RefreshToken{
		ID:        newCookieID,
		Token:     hashedRefresh,
		UserAgent: shared.GetMetadata(ctx, shared.KeyUserAgent),
		IPAddress: shared.GetMetadata(ctx, shared.KeyIPAddress),
		ExpiresAt: refreshExpiry.Unix(),
		UserID:    claimInfo.UserID,
	}

	return dAuth.TokenPair{
		AccessTokenStr: signedAccess,
		RefreshToken:   refreshToken,
	}, nil
}

func (j *JWTProvider) ValidateToken(ctx context.Context, tokenType dAuth.TokenType, tokenString string) (dAuth.ClaimInfo, error) {
	rawToken := tokenString

	if tokenType == dAuth.TypeRefresh {
		decrypted, err := j.enc.DecryptAES(ctx, EncRefreshToken, tokenString)
		if err != nil {
			j.logger.Error(ctx, shared.System, err.Error())
			return dAuth.ClaimInfo{}, dAuth.ErrTokenInvalid
		}
		rawToken = decrypted
	}

	token, err := jwt.ParseWithClaims(rawToken, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret.Jwt), nil
	})

	if err != nil || !token.Valid {
		return dAuth.ClaimInfo{}, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return dAuth.ClaimInfo{}, fmt.Errorf("invalid claims type")
	}

	return dAuth.ClaimInfo{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (j *JWTProvider) signToken(info dAuth.ClaimInfo, expiry time.Time) (string, error) {
	claims := jwtClaims{
		ClaimInfo: info,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(j.config.Secret.Jwt))
}
