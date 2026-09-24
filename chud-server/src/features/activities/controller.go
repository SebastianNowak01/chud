package activities

import (
	"context"
	"net/http"

	"github.com/sebnow/chud/platform/apperr"
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

func (c *ActivityAPIController) DeleteActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	actor, id, ok := actorAndID(w, r)
	if !ok {
		return
	}
	if err := c.service.DeleteActivity(ctx, id, actor); err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ActivityAPIController) ArchiveActivityHandler(w http.ResponseWriter, r *http.Request) {
	c.respondManaged(w, r, c.service.ArchiveActivity)
}

func (c *ActivityAPIController) RestoreActivityHandler(w http.ResponseWriter, r *http.Request) {
	c.respondManaged(w, r, c.service.RestoreActivity)
}

func (c *ActivityAPIController) respondManaged(
	w http.ResponseWriter,
	r *http.Request,
	action func(ctx context.Context, id string, actor Actor) (*Activity, *apperr.ServiceError),
) {
	ctx := r.Context()
	actor, id, ok := actorAndID(w, r)
	if !ok {
		return
	}
	activity, err := action(ctx, id, actor)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, activity)
}

func actorAndID(w http.ResponseWriter, r *http.Request) (Actor, string, bool) {
	claims, ok := auth.ExtractUserOrRespond(r.Context(), w, r)
	if !ok {
		return Actor{}, "", false
	}
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return Actor{}, "", false
	}
	return Actor{UserID: claims.UserID, IsAdmin: claims.IsAdmin}, id, true
}
