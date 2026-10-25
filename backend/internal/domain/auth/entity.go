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

type TokenProvider interface {
	GenerateTokenPair(ctx context.Context, info shared.ClaimInfo) (TokenPair, error)
	ValidateToken(ctx context.Context, tokenType shared.TokenType, token string) (shared.ClaimInfo, error)
}

var (
	SuccessRegister           shared.SuccessKey = "AUTH.REGISTER_SUCCESS"
	SuccessLogin              shared.SuccessKey = "AUTH.LOGIN_SUCCESS"
	SuccessLogout             shared.SuccessKey = "AUTH.LOGOUT_SUCCESS"
	SuccessRefresh            shared.SuccessKey = "AUTH.REFRESH_SUCCESS"
	SuccessVerificationSent   shared.SuccessKey = "AUTH.VERIFICATION_CODE_SENT"
	SuccessEmailVerified      shared.SuccessKey = "AUTH.EMAIL_VERIFIED"
	SuccessPasswordResetEmail shared.SuccessKey = "AUTH.PASSWORD_RESET_EMAIL_SENT"
	SuccessPasswordReset      shared.SuccessKey = "AUTH.PASSWORD_RESET_SUCCESS"
)
