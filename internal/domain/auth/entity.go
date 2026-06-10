package auth

import (
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
