package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sebnow/chud/platform/httpx"
)

func JwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString := tokenFromRequest(r)
		if tokenString == "" {
			httpx.RespondError(ctx, w, http.StatusUnauthorized, fmt.Errorf("brak tokenu, zaloguj się"))
			return
		}

		userClaims, err := ValidateJwt(tokenString)
		if err != nil {
			httpx.RespondError(ctx, w, http.StatusUnauthorized, fmt.Errorf("nieprawidłowy token, zaloguj się ponownie"))
			return
		}

		next.ServeHTTP(w, r.WithContext(SetUserInContext(ctx, userClaims)))
	})
}

const LoginTokenCookie = "LOGIN_TOKEN"

func tokenFromRequest(r *http.Request) string {
	if token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); found {
		return token
	}
	if cookie, err := r.Cookie(LoginTokenCookie); err == nil {
		return cookie.Value
	}
	return ""
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ExtractUserOrRespond(ctx, w, r)
		if !ok {
			return
		}
		if !user.IsAdmin {
			httpx.RespondError(ctx, w, http.StatusForbidden, fmt.Errorf("wymagane uprawnienia administratora"))
			return
		}
		next(w, r)
	}
}
