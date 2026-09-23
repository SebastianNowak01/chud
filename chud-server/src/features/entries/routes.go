package entries

import "net/http"

func RegisterRoutes(protectedRouter *http.ServeMux, c EntryAPIController) {
	protectedRouter.HandleFunc("GET /api/v1/activities/{id}/entries", c.GetEntriesByActivityHandler)
	protectedRouter.HandleFunc("POST /api/v1/activities/{id}/entries", c.CreateEntryHandler)
	protectedRouter.HandleFunc("GET /api/v1/entries", c.GetEntriesInRangeHandler)
	protectedRouter.HandleFunc("GET /api/v1/me/entries", c.GetMyEntriesInRangeHandler)
	protectedRouter.HandleFunc("GET /api/v1/users/{id}/entries", c.GetUserEntriesInRangeHandler)
	protectedRouter.HandleFunc("GET /api/v1/entries/{id}/media", c.GetMediaByEntryHandler)
	protectedRouter.HandleFunc("GET /api/v1/media/{id}", c.GetMediaHandler)
}
