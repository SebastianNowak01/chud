package stats

import (
	"net/http"

	"github.com/sebnow/chud/platform/httpx"
)

type StatsAPIController struct {
	service IStatsService
}

func NewStatsAPIController(service IStatsService) StatsAPIController {
	return StatsAPIController{service: service}
}

func (c *StatsAPIController) GetOccurrencesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from, to, err := httpx.DateRangeQuery(r)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	occurrences, svcErr := c.service.GetOccurrences(ctx, from, to)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, occurrences)
}

func (c *StatsAPIController) GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from, to, err := httpx.DateRangeQuery(r)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	leaderboard, svcErr := c.service.GetLeaderboard(ctx, from, to)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, leaderboard)
}
