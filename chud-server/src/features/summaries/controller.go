package summaries

import (
	"net/http"

	"github.com/sebnow/chud/platform/httpx"
)

type SummaryAPIController struct {
	service ISummaryService
}

func NewSummaryAPIController(service ISummaryService) SummaryAPIController {
	return SummaryAPIController{service: service}
}

func (c *SummaryAPIController) GetWeekSummaryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	week, err := httpx.DateQuery(r, "week")
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	summary, svcErr := c.service.Get(ctx, week)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, summary)
}

func (c *SummaryAPIController) RegenerateWeekSummaryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	week, err := httpx.DateQuery(r, "week")
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	summary, svcErr := c.service.Regenerate(ctx, week)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusAccepted, summary)
}
