package auth

import (
	aAuth "agropal/internal/app/auth"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
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

// ToVerification validates the email/code pair shared by the
// register/verify and login/verify endpoints.
func ToVerification(req request.VerificationRequest) (email, code string, err error) {
	if req.Email == nil || !utils.ValidEmail(*req.Email) {
		return "", "", dAuth.ErrInvalidEmail
	}
	if req.Code == nil || *req.Code == "" {
		return "", "", dAuth.ErrVerificationCodeInvalid
	}

	return *req.Email, *req.Code, nil
}

// ToResendVerification validates the email/purpose pair for the
// {purpose}/resend-code endpoint. `login` is deliberately not a resendable
// purpose here: unlike Login, this endpoint has no password check in front
// of it, so allowing purpose=login would let anyone trigger a login OTP
// email for any address with no credentials at all. A stuck login should
// just call Login again, which resends under the same cooldown.
func ToResendVerification(req request.ResendVerificationRequest, purposeStr string) (email string, purpose aAuth.VerificationPurpose, err error) {
	if req.Email == nil || !utils.ValidEmail(*req.Email) {
		return "", "", dAuth.ErrInvalidEmail
	}

	switch purposeStr {
	case string(aAuth.VerificationPurposeRegister):
		purpose = aAuth.VerificationPurposeRegister
	case string(aAuth.VerificationPurposeResetPassword):
		purpose = aAuth.VerificationPurposeResetPassword
	default:
		return "", "", shared.ErrInvalidRequest
	}

	return *req.Email, purpose, nil
}

// ToPasswordResetEmail validates the email for the reset-password/request
// endpoint.
func ToPasswordResetEmail(req request.PasswordResetRequest) (string, error) {
	if req.Email == nil || !utils.ValidEmail(*req.Email) {
		return "", dAuth.ErrInvalidEmail
	}

	return *req.Email, nil
}

// ToPasswordResetConfirmation validates the email/code/new-password triple
// for the reset-password/confirm endpoint.
func ToPasswordResetConfirmation(req request.PasswordResetConfirmRequest) (email, code, newPassword string, err error) {
	if req.Email == nil || !utils.ValidEmail(*req.Email) {
		return "", "", "", dAuth.ErrInvalidEmail
	}
	if req.Code == nil || *req.Code == "" {
		return "", "", "", dAuth.ErrVerificationCodeInvalid
	}
	if req.NewPassword == nil {
		return "", "", "", dAuth.ErrInvalidPassword
	}
	if err := utils.ValidPassword(*req.NewPassword); err != nil {
		return "", "", "", err
	}

	return *req.Email, *req.Code, *req.NewPassword, nil
}
