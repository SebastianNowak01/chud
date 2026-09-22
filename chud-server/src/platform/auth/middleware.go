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

		authHeader := r.Header.Get("Authorization")
		tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found || tokenString == "" {
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
