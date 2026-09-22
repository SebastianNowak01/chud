package users

import (
	"context"
	"net/http"
	"testing"

	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) *UserService {
	t.Helper()
	t.Setenv(config.JwtSecret, "test-secret")
	t.Setenv(config.JwtExpiryHours, "1")

	svc := NewUserService(UserServiceDeps{UserDAO: newFakeUserDAO()})
	require.NoError(t, svc.EnsureAdminUserExists(context.Background(), "admin", "admin-pass"))
	return svc
}

func findAdmin(t *testing.T, svc *UserService) UserResponse {
	t.Helper()
	users, svcErr := svc.GetAllUsers(context.Background())
	require.Nil(t, svcErr)
	for _, u := range users {
		if u.IsAdmin {
			return u
		}
	}
	t.Fatal("admin not found")
	return UserResponse{}
}

func TestEnsureAdminUserExists(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	users, svcErr := svc.GetAllUsers(ctx)
	require.Nil(t, svcErr)
	require.Len(t, users, 1)
	assert.Equal(t, "admin", users[0].Username)
	assert.True(t, users[0].IsAdmin)

	// Env always wins: re-running resets the password without duplicating the user.
	require.NoError(t, svc.EnsureAdminUserExists(ctx, "admin", "new-pass"))
	users, svcErr = svc.GetAllUsers(ctx)
	require.Nil(t, svcErr)
	assert.Len(t, users, 1)

	_, svcErr = svc.Login(ctx, LoginPayload{Username: "admin", Password: "admin-pass"})
	require.NotNil(t, svcErr)
	_, svcErr = svc.Login(ctx, LoginPayload{Username: "admin", Password: "new-pass"})
	assert.Nil(t, svcErr)
}

func TestEnsureAdminUserExistsKeepsSingleAdmin(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	// ADMIN_USER changed between restarts: the old admin becomes a regular user.
	require.NoError(t, svc.EnsureAdminUserExists(ctx, "root", "root-pass"))

	users, svcErr := svc.GetAllUsers(ctx)
	require.Nil(t, svcErr)
	admins := 0
	for _, u := range users {
		if u.IsAdmin {
			admins++
			assert.Equal(t, "root", u.Username)
		}
	}
	assert.Equal(t, 1, admins)
}

func TestMalformedIDIsNotFound(t *testing.T) {
	svc := newTestService(t)

	_, svcErr := svc.GetUser(context.Background(), "not-a-uuid")
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusNotFound, svcErr.Code)
}

func TestUserCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	created, svcErr := svc.CreateUser(ctx, CreateUserPayload{Username: "  alice ", Password: "secret1"})
	require.Nil(t, svcErr)
	assert.Equal(t, "alice", created.Username)
	assert.False(t, created.IsAdmin)

	got, svcErr := svc.GetUser(ctx, created.ID)
	require.Nil(t, svcErr)
	assert.Equal(t, created.ID, got.ID)

	updated, svcErr := svc.UpdateUser(ctx, created.ID, UpdateUserPayload{Username: "alicja"})
	require.Nil(t, svcErr)
	assert.Equal(t, "alicja", updated.Username)

	// Empty password on update keeps the old one.
	_, svcErr = svc.Login(ctx, LoginPayload{Username: "alicja", Password: "secret1"})
	require.Nil(t, svcErr)

	_, svcErr = svc.UpdateUser(ctx, created.ID, UpdateUserPayload{Username: "alicja", Password: "secret2"})
	require.Nil(t, svcErr)
	_, svcErr = svc.Login(ctx, LoginPayload{Username: "alicja", Password: "secret2"})
	require.Nil(t, svcErr)

	require.Nil(t, svc.DeleteUser(ctx, created.ID))
	_, svcErr = svc.GetUser(ctx, created.ID)
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusNotFound, svcErr.Code)
}

func TestCreateUserValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tests := []struct {
		name     string
		payload  CreateUserPayload
		wantCode int
	}{
		{"short username", CreateUserPayload{Username: "ab", Password: "secret1"}, http.StatusBadRequest},
		{"short password", CreateUserPayload{Username: "alice", Password: "123"}, http.StatusBadRequest},
		{"admin username taken", CreateUserPayload{Username: "admin", Password: "secret1"}, http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, svcErr := svc.CreateUser(ctx, tt.payload)
			require.NotNil(t, svcErr)
			assert.Equal(t, tt.wantCode, svcErr.Code)
		})
	}
}

func TestUpdateToTakenUsernameConflicts(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	_, svcErr := svc.CreateUser(ctx, CreateUserPayload{Username: "alice", Password: "secret1"})
	require.Nil(t, svcErr)
	bob, svcErr := svc.CreateUser(ctx, CreateUserPayload{Username: "bob", Password: "secret1"})
	require.Nil(t, svcErr)

	_, svcErr = svc.UpdateUser(ctx, bob.ID, UpdateUserPayload{Username: "alice"})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusConflict, svcErr.Code)
}

func TestAdminIsLocked(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	admin := findAdmin(t, svc)

	_, svcErr := svc.UpdateUser(ctx, admin.ID, UpdateUserPayload{Username: "hacker"})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusForbidden, svcErr.Code)

	svcErr = svc.DeleteUser(ctx, admin.ID)
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusForbidden, svcErr.Code)
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	_, svcErr := svc.CreateUser(ctx, CreateUserPayload{Username: "alice", Password: "secret1"})
	require.Nil(t, svcErr)

	adminLogin, svcErr := svc.Login(ctx, LoginPayload{Username: "admin", Password: "admin-pass"})
	require.Nil(t, svcErr)
	claims, err := auth.ValidateJwt(adminLogin.Token)
	require.NoError(t, err)
	assert.True(t, claims.IsAdmin)

	userLogin, svcErr := svc.Login(ctx, LoginPayload{Username: "alice", Password: "secret1"})
	require.Nil(t, svcErr)
	claims, err = auth.ValidateJwt(userLogin.Token)
	require.NoError(t, err)
	assert.Equal(t, "alice", claims.Username)
	assert.False(t, claims.IsAdmin)

	_, svcErr = svc.Login(ctx, LoginPayload{Username: "alice", Password: "wrong"})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusUnauthorized, svcErr.Code)

	_, svcErr = svc.Login(ctx, LoginPayload{Username: "nobody", Password: "secret1"})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusUnauthorized, svcErr.Code)
}
