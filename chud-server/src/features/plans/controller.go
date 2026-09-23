package plans

import (
	"net/http"

	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/httpx"
)

type PlanAPIController struct {
	service IPlanService
}

func NewPlanAPIController(service IPlanService) PlanAPIController {
	return PlanAPIController{
		service: service,
	}
}

func (c *PlanAPIController) GetPlansByActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	plans, err := c.service.GetPlansByActivity(ctx, id)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, plans)
}

func (c *PlanAPIController) CreatePlanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	var payload PlanPayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	plan, err := c.service.CreatePlan(ctx, id, claims.UserID, payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusCreated, plan)
}

func (c *PlanAPIController) UpdatePlanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	var payload PlanUpdatePayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	plan, err := c.service.UpdatePlan(ctx, id, claims.UserID, payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, plan)
}

func (c *PlanAPIController) DeletePlanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	if err := c.service.DeletePlan(ctx, id, claims.UserID); err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
