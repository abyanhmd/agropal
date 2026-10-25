package auth

import (
	dAuth "agropal/internal/domain/auth"
	"net/http"
	"time"
)

const authCookieName = "rid"

func setAuthCookie(w http.ResponseWriter, tokenPair dAuth.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    tokenPair.RefreshToken.ID,
		Expires:  time.Unix(tokenPair.RefreshToken.ExpiresAt, 0),
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func removeAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Expires:  time.Now().UTC(),
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
