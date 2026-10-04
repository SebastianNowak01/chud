package summaries

import "net/http"

func RegisterRoutes(protectedRouter *http.ServeMux, c SummaryAPIController) {
	protectedRouter.HandleFunc("GET /api/v1/summaries/week", c.GetWeekSummaryHandler)
	protectedRouter.HandleFunc("POST /api/v1/summaries/week/regenerate", c.RegenerateWeekSummaryHandler)
}
