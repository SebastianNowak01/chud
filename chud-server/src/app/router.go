package app

import (
	"embed"
	"net/http"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/plans"
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

	// Activities, plans and entries
	activities.RegisterRoutes(protectedRouter, h.Activity)
	plans.RegisterRoutes(protectedRouter, h.Plan)
	entries.RegisterRoutes(protectedRouter, h.Entry)

	// Health
	publicRouter.HandleFunc("GET /api/v1/health", GetHealthCheckHandler)

	// SPA Handler for UI
	publicRouter.Handle("/", newSPAHandler(staticFiles))
}
