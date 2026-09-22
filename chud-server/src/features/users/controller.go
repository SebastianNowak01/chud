package users

import (
	"net/http"

	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/httpx"
)

type UserAPIController struct {
	service IUserService
}

func NewUserAPIController(service IUserService) UserAPIController {
	return UserAPIController{
		service: service,
	}
}

func (c *UserAPIController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload LoginPayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	loginResponse, err := c.service.Login(ctx, payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, loginResponse)
}

func (c *UserAPIController) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	user, err := c.service.GetUser(ctx, claims.UserID)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, user)
}

func (c *UserAPIController) UpdateMeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	var payload UpdateMePayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	user, err := c.service.UpdateMe(ctx, claims.UserID, payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, user)
}

func (c *UserAPIController) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := c.service.GetAllUsers(ctx)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, users)
}

func (c *UserAPIController) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := c.service.GetUser(ctx, r.PathValue("id"))
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, user)
}

func (c *UserAPIController) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload CreateUserPayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	user, err := c.service.CreateUser(ctx, payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusCreated, user)
}

func (c *UserAPIController) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload UpdateUserPayload
	if !httpx.DecodeJSONOrRespond(ctx, w, r, &payload) {
		return
	}

	user, err := c.service.UpdateUser(ctx, r.PathValue("id"), payload)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, user)
}

func (c *UserAPIController) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := c.service.DeleteUser(ctx, r.PathValue("id")); err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
