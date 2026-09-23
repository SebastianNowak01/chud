package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sebnow/chud/platform/httpx"
	"github.com/sebnow/chud/platform/log"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

func SetUserInContext(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

func GetUserClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	return claims, ok
}

func ExtractUserOrRespond(ctx context.Context, w http.ResponseWriter, _ *http.Request) (*UserClaims, bool) {
	logger := log.FromContext(ctx)
	user, ok := GetUserClaimsFromContext(ctx)
	if !ok {
		logger.Warn().Msg("User claims not found in context")
		httpx.RespondError(ctx, w, http.StatusUnauthorized, fmt.Errorf("brak zalogowanego użytkownika"))
		return nil, false
	}
	return user, true
}
