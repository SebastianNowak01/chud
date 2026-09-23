package app

import (
	"embed"
	"net/http"

	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/users"
)

func SetupRouters(
	publicRouter *http.ServeMux,
	protectedRouter *http.ServeMux,
	staticFiles embed.FS,
	h Handlers,
) {
	users.RegisterRoutes(publicRouter, protectedRouter, h.User)

	activities.RegisterRoutes(protectedRouter, h.Activity)
	plans.RegisterRoutes(protectedRouter, h.Plan)
	entries.RegisterRoutes(protectedRouter, h.Entry)
	stats.RegisterRoutes(protectedRouter, h.Stats)

	publicRouter.HandleFunc("GET /api/v1/health", GetHealthCheckHandler)

	publicRouter.Handle("/", newSPAHandler(staticFiles))
}
