package stats

import "net/http"

func RegisterRoutes(protectedRouter *http.ServeMux, c StatsAPIController) {
	protectedRouter.HandleFunc("GET /api/v1/stats", c.GetLeaderboardHandler)
	protectedRouter.HandleFunc("GET /api/v1/occurrences", c.GetOccurrencesHandler)
}
