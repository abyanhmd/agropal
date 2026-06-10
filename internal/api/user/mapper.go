package user

import (
	dUser "agropal/internal/domain/user"
	"agropal/internal/response"
)

func DomainToUserResponse(domain dUser.User) response.UserResponse {
	return response.UserResponse{
		UserID:   domain.ID,
		Name:     domain.Name,
		Email:    domain.Email,
		Username: domain.Username,
		Avatar:   domain.Avatar,
		Password: domain.Password,
		Role:     domain.Role,
	}
}
