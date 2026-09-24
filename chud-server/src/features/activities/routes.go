package activities

import "net/http"

func RegisterRoutes(protectedRouter *http.ServeMux, c ActivityAPIController) {
	protectedRouter.HandleFunc("GET /api/v1/activities", c.GetAllActivitiesHandler)
	protectedRouter.HandleFunc("POST /api/v1/activities", c.CreateActivityHandler)
	protectedRouter.HandleFunc("GET /api/v1/activities/{id}", c.GetActivityHandler)
	protectedRouter.HandleFunc("GET /api/v1/activities/{id}/members", c.GetMembersHandler)
	protectedRouter.HandleFunc("DELETE /api/v1/activities/{id}", c.DeleteActivityHandler)
	protectedRouter.HandleFunc("POST /api/v1/activities/{id}/archive", c.ArchiveActivityHandler)
	protectedRouter.HandleFunc("POST /api/v1/activities/{id}/restore", c.RestoreActivityHandler)
}
