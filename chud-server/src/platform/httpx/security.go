package httpx

import (
	"net/http"

	"github.com/sebnow/chud/platform/constants"
)

const appContentSecurityPolicy = "default-src 'self'; img-src 'self' blob: data:; media-src 'self' blob:; " +
	"style-src 'self' 'unsafe-inline'; font-src 'self' data:; object-src 'none'; base-uri 'none'; " +
	"frame-ancestors 'none'; form-action 'self'"

const MediaContentSecurityPolicy = "default-src 'none'; img-src 'self'; media-src 'self'; style-src 'unsafe-inline'; sandbox"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set(constants.HTTPHeaderContentTypeOptions, "nosniff")
		h.Set(constants.HTTPHeaderFrameOptions, "DENY")
		h.Set(constants.HTTPHeaderReferrerPolicy, "same-origin")
		h.Set(constants.HTTPHeaderContentSecurityPolicy, appContentSecurityPolicy)
		next.ServeHTTP(w, r)
	})
}
