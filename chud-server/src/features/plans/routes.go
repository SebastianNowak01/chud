package plans

import "net/http"

// RegisterRoutes registers the plan API routes.
func RegisterRoutes(protectedRouter *http.ServeMux, c PlanAPIController) {
	protectedRouter.HandleFunc("GET /api/v1/activities/{id}/plans", c.GetPlansByActivityHandler)
	protectedRouter.HandleFunc("POST /api/v1/activities/{id}/plans", c.CreatePlanHandler)
	protectedRouter.HandleFunc("PUT /api/v1/plans/{id}", c.UpdatePlanHandler)
	protectedRouter.HandleFunc("DELETE /api/v1/plans/{id}", c.DeletePlanHandler)
}
