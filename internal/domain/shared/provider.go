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
	UserID uuid.UUID `json:"sub"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}
