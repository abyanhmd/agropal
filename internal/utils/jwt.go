package utils

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Logger *Logger
	Config *config.Config
	Enc    *Encryption
}

func NewJWT(logger *Logger, config *config.Config, enc *Encryption) *JWT {
	return &JWT{
		Logger: logger,
		Config: config,
		Enc:    enc,
	}
}

type TokenType int

const (
	AccessToken TokenType = iota
	RefreshToken
)

type ClaimInfo struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type Claims struct {
	ClaimInfo
	jwt.RegisteredClaims
}

func (j *JWT) GenerateToken(ctx context.Context, tokenType TokenType, claimInfo ClaimInfo) (string, error) {
	now := time.Now()
	var expiredAt time.Time

	switch tokenType {
	case AccessToken:
		duration := time.Duration(30)
		expiredAt = now.Add(duration * time.Minute)
	case RefreshToken:
		duration := time.Duration(7)
		expiredAt = now.Add(duration * 24 * time.Hour)
	default:
		return "", nil
	}

	claims := Claims{
		ClaimInfo: claimInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString([]byte(j.Config.Secret.Jwt))
	if signErr != nil {
		j.Logger.Error(ctx, shared.System, signErr.Error())
		return "", fmt.Errorf("failed to sign token. %w", signErr)
	}

	if tokenType == RefreshToken {
		return j.Enc.EncryptAES(ctx, EncRefreshToken, tokenString)
	}

	return tokenString, nil
}

func (j *JWT) ParseToken(ctx context.Context, tokenType TokenType, tokenString string) (*Claims, error) {
	if tokenType == RefreshToken {
		temp, err := j.Enc.EncryptAES(ctx, EncRefreshToken, tokenString)
		if err != nil {
			return nil, err
		}
		tokenString = temp
	}

	token, parseErr := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			j.Logger.Error(ctx, shared.System, "unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Config.Secret.Jwt), nil
	})

	if parseErr != nil {
		j.Logger.Error(ctx, shared.System, parseErr.Error())
		return nil, fmt.Errorf("failed to parse token. %w", parseErr)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
