package auth

import (
	"agropal/internal/utils"

	"github.com/google/uuid"
)

type AuthToken struct {
	AccessToken  string
	RefreshToken string
	Claims       *utils.Claims
}

type RefreshToken struct {
	ID        string
	Token     string
	UserAgent string
	IPAddress string
	UserID    uuid.UUID
}
