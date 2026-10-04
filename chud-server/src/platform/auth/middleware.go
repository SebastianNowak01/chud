package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sebnow/chud/platform/httpx"
)

type Session struct {
	Version int
	IsAdmin bool
}

type SessionLookup func(ctx context.Context, userID string) (*Session, error)

func JwtAuth(lookup SessionLookup, next http.Handler) http.Handler {
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

		session, err := lookup(ctx, userClaims.UserID)
		if err != nil {
			httpx.RespondError(ctx, w, http.StatusInternalServerError, err)
			return
		}
		if session == nil || session.Version != userClaims.Version {
			httpx.RespondError(ctx, w, http.StatusUnauthorized, fmt.Errorf("sesja wygasła, zaloguj się ponownie"))
			return
		}
		userClaims.IsAdmin = session.IsAdmin

		next.ServeHTTP(w, r.WithContext(SetUserInContext(ctx, userClaims)))
	})
}

const LoginTokenCookie = "LOGIN_TOKEN"

func tokenFromRequest(r *http.Request) string {
	if token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); found {
		return token
	}
	if !isMediaRequest(r) {
		return ""
	}
	if cookie, err := r.Cookie(LoginTokenCookie); err == nil {
		return cookie.Value
	}
	return ""
}

func isMediaRequest(r *http.Request) bool {
	return (r.Method == http.MethodGet || r.Method == http.MethodHead) && strings.HasPrefix(r.URL.Path, "/api/v1/media/")
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
