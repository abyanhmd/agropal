package auth

import (
	"context"

	"github.com/google/uuid"
)

type TokenType int

const (
	TypeAccess TokenType = iota
	TypeRefresh
)

type ClaimInfo struct {
	UserID uuid.UUID `json:"sub"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

type TokenProvider interface {
	GenerateTokenPair(ctx context.Context, info ClaimInfo) (TokenPair, error)
	ValidateToken(ctx context.Context, tokenType TokenType, token string) (ClaimInfo, error)
}
