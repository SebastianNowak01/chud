package users

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	minUsernameLength = 3
	maxUsernameLength = 32
	minPasswordLength = 6
	maxPasswordLength = 72 // bcrypt limit
)

type IUserService interface {
	GetAllUsers(ctx context.Context) ([]UserResponse, *apperr.ServiceError)
	GetUser(ctx context.Context, id string) (*UserResponse, *apperr.ServiceError)
	CreateUser(ctx context.Context, payload CreateUserPayload) (*UserResponse, *apperr.ServiceError)
	UpdateUser(ctx context.Context, id string, payload UpdateUserPayload) (*UserResponse, *apperr.ServiceError)
	DeleteUser(ctx context.Context, id string) *apperr.ServiceError
	Login(ctx context.Context, payload LoginPayload) (*LoginResponse, *apperr.ServiceError)
	EnsureAdminUserExists(ctx context.Context, username, password string) error
}

type UserServiceDeps struct {
	UserDAO IUserDAO
}

type UserService struct {
	dao IUserDAO
	// dummyHash is compared against when a login username does not exist, so response time does not reveal it.
	dummyHash []byte
}

func NewUserService(deps UserServiceDeps) *UserService {
	dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	return &UserService{dao: deps.UserDAO, dummyHash: dummyHash}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]UserResponse, *apperr.ServiceError) {
	users, err := s.dao.GetAllUsers(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	result := make([]UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, u.ToResponse())
	}
	return result, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*UserResponse, *apperr.ServiceError) {
	user, svcErr := s.getUser(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}
	response := user.ToResponse()
	return &response, nil
}

func (s *UserService) CreateUser(ctx context.Context, payload CreateUserPayload) (*UserResponse, *apperr.ServiceError) {
	username := strings.TrimSpace(payload.Username)
	if svcErr := validateUsername(username); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := validatePassword(payload.Password); svcErr != nil {
		return nil, svcErr
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.NewInternalError("failed to hash password: %w", err)
	}

	created, err := s.dao.InsertUser(ctx, &User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, daoError(err)
	}

	log.FromContext(ctx).Info().Str("username", username).Msg("User created")
	response := created.ToResponse()
	return &response, nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	id string,
	payload UpdateUserPayload,
) (*UserResponse, *apperr.ServiceError) {
	user, svcErr := s.getUser(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}
	if user.IsAdmin {
		return nil, apperr.NewForbiddenError("the admin user is managed through environment variables")
	}

	username := strings.TrimSpace(payload.Username)
	if svcErr := validateUsername(username); svcErr != nil {
		return nil, svcErr
	}
	user.Username = username

	if payload.Password != "" {
		if svcErr := validatePassword(payload.Password); svcErr != nil {
			return nil, svcErr
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperr.NewInternalError("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hash)
	}

	updated, err := s.dao.UpdateUser(ctx, user)
	if err != nil {
		return nil, daoError(err)
	}

	log.FromContext(ctx).Info().Str("username", username).Msg("User updated")
	response := updated.ToResponse()
	return &response, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) *apperr.ServiceError {
	user, svcErr := s.getUser(ctx, id)
	if svcErr != nil {
		return svcErr
	}
	if user.IsAdmin {
		return apperr.NewForbiddenError("the admin user cannot be deleted")
	}
	if err := s.dao.DeleteUser(ctx, id); err != nil {
		return daoError(err)
	}

	log.FromContext(ctx).Info().Str("username", user.Username).Msg("User deleted")
	return nil
}

func (s *UserService) Login(ctx context.Context, payload LoginPayload) (*LoginResponse, *apperr.ServiceError) {
	user, err := s.dao.GetUserByUsername(ctx, strings.TrimSpace(payload.Username))
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, daoError(err)
		}
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(payload.Password))
		return nil, apperr.NewUnauthorizedError("invalid username or password")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)) != nil {
		return nil, apperr.NewUnauthorizedError("invalid username or password")
	}

	token, err := auth.NewJwt(user.Username, user.IsAdmin)
	if err != nil {
		return nil, apperr.NewInternalError("failed to create token: %w", err)
	}

	return &LoginResponse{Token: token, User: user.ToResponse()}, nil
}

// EnsureAdminUserExists makes the env user the only admin, creating it or resetting its password,
// so the environment always wins.
func (s *UserService) EnsureAdminUserExists(ctx context.Context, username, password string) error {
	logger := log.FromContext(ctx)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.dao.DemoteAdminsExcept(ctx, username); err != nil {
		return err
	}

	existing, err := s.dao.GetUserByUsername(ctx, username)
	switch {
	case err == nil:
		existing.PasswordHash = string(hash)
		existing.IsAdmin = true
		if _, err := s.dao.UpdateUser(ctx, existing); err != nil {
			return err
		}
		logger.Info().Str("username", username).Msg("Admin user reset from environment")
	case errors.Is(err, db.ErrNotFound):
		_, err := s.dao.InsertUser(ctx, &User{
			ID:           uuid.NewString(),
			Username:     username,
			PasswordHash: string(hash),
			IsAdmin:      true,
		})
		if err != nil {
			return err
		}
		logger.Info().Str("username", username).Msg("Admin user created from environment")
	default:
		return err
	}

	return nil
}

// getUser loads a user by ID, treating malformed IDs as not found.
func (s *UserService) getUser(ctx context.Context, id string) (*User, *apperr.ServiceError) {
	if uuid.Validate(id) != nil {
		return nil, apperr.NewNotFoundError("user not found")
	}
	user, err := s.dao.GetUserByID(ctx, id)
	if err != nil {
		return nil, daoError(err)
	}
	return user, nil
}

func validateUsername(username string) *apperr.ServiceError {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return apperr.NewBadRequestError(
			"username must be between %d and %d characters",
			minUsernameLength,
			maxUsernameLength,
		)
	}
	return nil
}

func validatePassword(password string) *apperr.ServiceError {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return apperr.NewBadRequestError(
			"password must be between %d and %d characters",
			minPasswordLength,
			maxPasswordLength,
		)
	}
	return nil
}

func daoError(err error) *apperr.ServiceError {
	switch {
	case errors.Is(err, db.ErrNotFound):
		return apperr.NewNotFoundError("user not found")
	case errors.Is(err, db.ErrAlreadyExists):
		return apperr.NewConflictError("username already taken")
	default:
		return apperr.NewInternalError("%w", err)
	}
}
