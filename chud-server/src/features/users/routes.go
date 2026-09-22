package users

import (
	"net/http"

	"github.com/sebnow/chud/platform/auth"
)

// RegisterRoutes registers the user and auth API routes.
func RegisterRoutes(publicRouter *http.ServeMux, protectedRouter *http.ServeMux, c UserAPIController) {
	publicRouter.HandleFunc("POST /api/v1/auth/login", c.LoginHandler)

	protectedRouter.HandleFunc("GET /api/v1/me", c.GetMeHandler)
	protectedRouter.HandleFunc("PATCH /api/v1/me", c.UpdateMeHandler)

	protectedRouter.HandleFunc("GET /api/v1/users", c.GetAllUsersHandler)
	protectedRouter.HandleFunc("GET /api/v1/users/{id}", c.GetUserHandler)
	protectedRouter.HandleFunc("POST /api/v1/users", auth.RequireAdmin(c.CreateUserHandler))
	protectedRouter.HandleFunc("PUT /api/v1/users/{id}", auth.RequireAdmin(c.UpdateUserHandler))
	protectedRouter.HandleFunc("DELETE /api/v1/users/{id}", auth.RequireAdmin(c.DeleteUserHandler))
}
