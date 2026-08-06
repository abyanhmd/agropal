package auth

import (
	"agropal/internal/domain/shared"
	"context"

	"github.com/google/uuid"
)

type TokenPair struct {
	AccessTokenStr string
	RefreshToken   RefreshToken
}

type RefreshToken struct {
	ID        string
	Token     string
	UserAgent string
	IPAddress string
	ExpiresAt int64
	UserID    uuid.UUID
}

type Role string

const (
	RoleAdmin Role = "ADMIN" // System level
	RoleUser  Role = "USER"  // Default user
	RoleGuest Role = "GUEST" // Temporary/Viewer
)

type TokenProvider interface {
	GenerateTokenPair(ctx context.Context, info shared.ClaimInfo) (TokenPair, error)
	ValidateToken(ctx context.Context, tokenType shared.TokenType, token string) (shared.ClaimInfo, error)
}

var (
	SuccessRegister shared.SuccessKey = "AUTH.REGISTER_SUCCESS"
	SuccessLogin    shared.SuccessKey = "AUTH.LOGIN_SUCCESS"
	SuccessLogout   shared.SuccessKey = "AUTH.LOGOUT_SUCCESS"
	SuccessRefresh  shared.SuccessKey = "AUTH.REFRESH_SUCCESS"
)
