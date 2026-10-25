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

// accessTokenBody is what a sign-in response carries. The refresh token travels only in the httpOnly
// cookie; putting it in the body too would hand it to any script on the page.
type accessTokenBody struct {
	AccessTokenStr string
}

func bodyFor(tokenPair dAuth.TokenPair) accessTokenBody {
	return accessTokenBody{AccessTokenStr: tokenPair.AccessTokenStr}
}
