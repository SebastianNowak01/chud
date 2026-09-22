package app

import (
	"embed"
	"net/http"

	"github.com/sebnow/chud/features/users"
)

func SetupRouters(
	publicRouter *http.ServeMux,
	protectedRouter *http.ServeMux,
	staticFiles embed.FS,
	h Handlers,
) {
	// Users
	users.RegisterRoutes(publicRouter, protectedRouter, h.User)

	// Health
	publicRouter.HandleFunc("GET /api/v1/health", GetHealthCheckHandler)

	// SPA Handler for UI
	publicRouter.Handle("/", newSPAHandler(staticFiles))
}
