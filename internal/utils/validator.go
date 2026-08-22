package utils

import (
	dAuth "agropal/internal/domain/auth"
	"net/mail"
	"strings"
)

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidPassword(password string) error {
	if len(password) < 8 {
		return dAuth.ErrPasswordTooShort
	}
	if len(password) > 64 {
		return dAuth.ErrPasswordTooLong
	}

	var hasLower, hasUpper, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case 'a' <= c && c <= 'z':
			hasLower = true
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case '0' <= c && c <= '9':
			hasNumber = true
		case strings.ContainsRune("@$!%*?&", c):
			hasSpecial = true
		}
	}

	switch {
	case !hasLower:
		return dAuth.ErrPasswordHasNoLower
	case !hasUpper:
		return dAuth.ErrPasswordHasNoUpper
	case !hasNumber:
		return dAuth.ErrPasswordHasNoNumber
	case !hasSpecial:
		return dAuth.ErrPasswordHasNoSpecial
	}

	return nil
}
