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

	for _, c := range password {
		switch {
		case 'a' <= c && c <= 'z':
			return dAuth.ErrPasswordHasNoLower
		case 'A' <= c && c <= 'Z':
			return dAuth.ErrPasswordHasNoUpper
		case '0' <= c && c <= '9':
			return dAuth.ErrPasswordHasNoNumber
		case strings.ContainsRune("@$!%*?&", c):
			return dAuth.ErrPasswordHasNoSpecial
		}
	}

	return nil
}
