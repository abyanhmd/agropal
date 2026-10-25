package shared

import (
	"github.com/google/uuid"
)

type TokenType int

const (
	TypeAccess TokenType = iota
	TypeRefresh
)

type ClaimInfo struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}
