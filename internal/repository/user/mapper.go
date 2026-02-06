package user

import dUser "agropal/internal/domain/user"

func DomainToModelUser(domain dUser.User) ModelUser {
	return ModelUser{
		UserID:   domain.UserID,
		Name:     domain.Name,
		Email:    domain.Email,
		Username: domain.Username,
		Avatar:   domain.Avatar,
		Password: domain.Password,
		Role:     domain.Role,
	}
}

func (m ModelUser) ToDomain() dUser.User {
	return dUser.User{
		UserID:   m.UserID,
		Name:     m.Name,
		Email:    m.Email,
		Username: m.Username,
		Avatar:   m.Avatar,
		Password: m.Password,
		Role:     m.Role,
	}
}
