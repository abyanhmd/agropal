package auth

import dAuth "agropal/internal/domain/auth"

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
