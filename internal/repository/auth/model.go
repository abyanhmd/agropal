package auth

import "github.com/google/uuid"

type RefreshTokenModel struct {
	ID        string    `db:"refresh_token_id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	UserAgent string    `db:"user_agent"`
	IPAddress string    `db:"ip_address"`
}
