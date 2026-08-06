package auth

import (
	aAuth "agropal/internal/app/auth"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/request"
	"agropal/internal/utils"
)

func ToRegisterCommand(req request.UserRequest) (aAuth.CreateUserCommand, error) {
	// ----- EMAIL ----------
	if req.Email == nil {
		return aAuth.CreateUserCommand{}, dAuth.ErrInvalidEmail
	}
	if !utils.ValidEmail(*req.Email) {
		return aAuth.CreateUserCommand{}, dAuth.ErrInvalidEmail
	}

	// ----- PASSWORD ----------
	if req.Password == nil {
		return aAuth.CreateUserCommand{}, dAuth.ErrInvalidPassword
	}
	if err := utils.ValidPassword(*req.Password); err != nil {
		return aAuth.CreateUserCommand{}, err
	}

	// ----- USERNAME ----------
	if req.Username == nil {
		return aAuth.CreateUserCommand{}, dAuth.ErrInvalidUsername
	}

	// ----- NAME ----------
	if req.Name == nil {
		return aAuth.CreateUserCommand{}, dAuth.ErrInvalidName
	}

	return aAuth.CreateUserCommand{
		Name:     *req.Name,
		Email:    *req.Email,
		Username: *req.Username,
		Password: *req.Password,
		Avatar:   req.Avatar,
	}, nil
}

func ToLoginCommand(req request.UserRequest) (aAuth.LoginUserCommand, error) {
	if req.Email == nil && req.Username == nil {
		return aAuth.LoginUserCommand{}, dAuth.ErrInvalidCredentials
	}
	if req.Email != nil && !utils.ValidEmail(*req.Email) {
		return aAuth.LoginUserCommand{}, dAuth.ErrInvalidEmail
	}
	if req.Password == nil {
		return aAuth.LoginUserCommand{}, dAuth.ErrInvalidPassword
	}

	return aAuth.LoginUserCommand{
		Email:    req.Email,
		Username: req.Username,
		Password: *req.Password,
	}, nil
}
