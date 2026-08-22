package middleware

import (
	"net/http"
	"time"
)

const (
	SessionCookieName = "rid"
)

func SetSessionCookie(cookie string, expiryAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    cookie,
		Expires:  expiryAt,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func DeleteSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Expires:  time.Now().UTC(),
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
