package users

import (
	"net/http"

	"github.com/sebnow/chud/platform/auth"
)

// RegisterRoutes registers the user and auth API routes.
func RegisterRoutes(publicRouter *http.ServeMux, protectedRouter *http.ServeMux, c UserAPIController) {
	publicRouter.HandleFunc("POST /api/v1/auth/login", c.LoginHandler)
	protectedRouter.HandleFunc("GET /api/v1/auth/me", c.MeHandler)

	protectedRouter.HandleFunc("GET /api/v1/users", auth.RequireAdmin(c.GetAllUsersHandler))
	protectedRouter.HandleFunc("POST /api/v1/users", auth.RequireAdmin(c.CreateUserHandler))
	protectedRouter.HandleFunc("GET /api/v1/users/{id}", auth.RequireAdmin(c.GetUserHandler))
	protectedRouter.HandleFunc("PUT /api/v1/users/{id}", auth.RequireAdmin(c.UpdateUserHandler))
	protectedRouter.HandleFunc("DELETE /api/v1/users/{id}", auth.RequireAdmin(c.DeleteUserHandler))
}
