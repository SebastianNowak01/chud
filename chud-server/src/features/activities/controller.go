package activities

import (
	"net/http"

	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/httpx"
)

type ActivityAPIController struct {
	service IActivityService
}

func NewActivityAPIController(service IActivityService) ActivityAPIController {
	return ActivityAPIController{
		service: service,
	}
}

func (c *ActivityAPIController) GetAllActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	activities, err := c.service.GetAllActivities(ctx)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, activities)
}

func (c *ActivityAPIController) GetActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	activity, err := c.service.GetActivity(ctx, id)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, activity)
}

func (c *ActivityAPIController) CreateActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	var payload CreateActivityPayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	activity, err := c.service.CreateActivity(ctx, claims.UserID, payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusCreated, activity)
}

func (c *ActivityAPIController) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	members, err := c.service.GetMembers(ctx, id)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, members)
}
