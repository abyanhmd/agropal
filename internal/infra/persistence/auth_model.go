package persistence

import (
	dAuth "agropal/internal/domain/auth"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenModel struct {
	ID        string    `db:"refresh_token_id"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	Token     string    `db:"token"`
	UserAgent string    `db:"user_agent"`
	IPAddress string    `db:"ip_address"`
}

func ToRefreshTokenModel(domain dAuth.RefreshToken) RefreshTokenModel {
	return RefreshTokenModel{
		ID:        domain.ID,
		UserID:    domain.UserID,
		ExpiresAt: time.Unix(domain.ExpiresAt, 0).UTC(),
		Token:     domain.Token,
		UserAgent: domain.UserAgent,
		IPAddress: domain.IPAddress,
	}
}

func (m RefreshTokenModel) ToDomain() dAuth.RefreshToken {
	return dAuth.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		ExpiresAt: m.ExpiresAt.Unix(),
		Token:     m.Token,
		UserAgent: m.UserAgent,
		IPAddress: m.IPAddress,
	}
}
