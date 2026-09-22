package activities

import "net/http"

// RegisterRoutes registers the activity API routes.
func RegisterRoutes(protectedRouter *http.ServeMux, c ActivityAPIController) {
	protectedRouter.HandleFunc("GET /api/v1/activities", c.GetAllActivitiesHandler)
	protectedRouter.HandleFunc("POST /api/v1/activities", c.CreateActivityHandler)
	protectedRouter.HandleFunc("GET /api/v1/activities/{id}", c.GetActivityHandler)
	protectedRouter.HandleFunc("GET /api/v1/activities/{id}/members", c.GetMembersHandler)
}
