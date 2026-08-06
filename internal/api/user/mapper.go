package user

import (
	dUser "agropal/internal/domain/user"
	"agropal/internal/response"
)

func DomainToUserResponse(domain dUser.User) response.UserResponse {
	return response.UserResponse{
		UserID:   domain.ID,
		Name:     domain.Account.Name,
		Email:    domain.Account.Email,
		Username: domain.Account.Username,
		Avatar:   domain.Account.Avatar,
		Role:     string(domain.Security.Role),
	}
}
