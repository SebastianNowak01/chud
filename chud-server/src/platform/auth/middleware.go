package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sebnow/chud/platform/httpx"
)

// JwtAuth middleware validates JWT tokens from the Authorization header.
func JwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString := tokenFromRequest(r)
		if tokenString == "" {
			httpx.RespondError(ctx, w, http.StatusUnauthorized, fmt.Errorf("no token provided"))
			return
		}

		userClaims, err := ValidateJwt(tokenString)
		if err != nil {
			httpx.RespondError(ctx, w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
			return
		}

		next.ServeHTTP(w, r.WithContext(SetUserInContext(ctx, userClaims)))
	})
}

// LoginTokenCookie is set by the UI; it lets plain <img>/<video> requests authenticate.
const LoginTokenCookie = "LOGIN_TOKEN"

// tokenFromRequest reads the JWT from the Authorization header, falling back to the login cookie.
func tokenFromRequest(r *http.Request) string {
	if token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); found {
		return token
	}
	if cookie, err := r.Cookie(LoginTokenCookie); err == nil {
		return cookie.Value
	}
	return ""
}

// RequireAdmin allows the request only if the authenticated user is the admin.
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ExtractUserOrRespond(ctx, w, r)
		if !ok {
			return
		}
		if !user.IsAdmin {
			httpx.RespondError(ctx, w, http.StatusForbidden, fmt.Errorf("admin access required"))
			return
		}
		next(w, r)
	}
}
