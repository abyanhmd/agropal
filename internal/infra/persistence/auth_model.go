package persistence

import (
	dAuth "agropal/internal/domain/auth"

	"github.com/google/uuid"
)

type RefreshTokenModel struct {
	ID        string    `db:"refresh_token_id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	UserAgent string    `db:"user_agent"`
	IPAddress string    `db:"ip_address"`
}

func ToRefreshTokenModel(domain dAuth.RefreshToken) RefreshTokenModel {
	return RefreshTokenModel{
		ID:        domain.ID,
		UserID:    domain.UserID,
		Token:     domain.Token,
		UserAgent: domain.UserAgent,
		IPAddress: domain.IPAddress,
	}
}

func (m RefreshTokenModel) ToDomain() dAuth.RefreshToken {
	return dAuth.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		UserAgent: m.UserAgent,
		IPAddress: m.IPAddress,
	}
}
