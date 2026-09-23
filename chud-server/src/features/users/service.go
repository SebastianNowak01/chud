package users

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/log"
	"github.com/sebnow/chud/platform/ratelimit"
	"golang.org/x/crypto/bcrypt"
)

const (
	minUsernameLength = 3
	maxUsernameLength = 32
	minPasswordLength = 6
	maxPasswordLength = 72
	maxLoginFailures  = 10
	loginFailWindow   = 15 * time.Minute
)

type IUserService interface {
	GetAllUsers(ctx context.Context) ([]User, *apperr.ServiceError)
	GetUser(ctx context.Context, id string) (*User, *apperr.ServiceError)
	CreateUser(ctx context.Context, payload CreateUserPayload) (*User, *apperr.ServiceError)
	UpdateUser(ctx context.Context, id string, payload UpdateUserPayload) (*User, *apperr.ServiceError)
	UpdateMe(ctx context.Context, id string, payload UpdateMePayload) (*User, *apperr.ServiceError)
	DeleteUser(ctx context.Context, id string) *apperr.ServiceError
	Login(ctx context.Context, payload LoginPayload) (*LoginResponse, *apperr.ServiceError)
	EnsureAdminUserExists(ctx context.Context, username, password string) error
	IsAdmin(ctx context.Context, id string) (bool, error)
}

type UserServiceDeps struct {
	UserDAO IUserDAO
}

type UserService struct {
	dao          IUserDAO
	dummyHash    []byte
	loginLimiter *ratelimit.Limiter
}

func NewUserService(deps UserServiceDeps) *UserService {
	dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	return &UserService{
		dao:          deps.UserDAO,
		dummyHash:    dummyHash,
		loginLimiter: ratelimit.New(maxLoginFailures, loginFailWindow),
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]User, *apperr.ServiceError) {
	users, err := s.dao.GetAllUsers(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	return users, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, *apperr.ServiceError) {
	return s.getUser(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, payload CreateUserPayload) (*User, *apperr.ServiceError) {
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
		Color:        randomColor(),
	})
	if err != nil {
		return nil, daoError(err)
	}

	log.FromContext(ctx).Info().Str("username", username).Msg("User created")
	return created, nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	id string,
	payload UpdateUserPayload,
) (*User, *apperr.ServiceError) {
	user, svcErr := s.getUser(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}
	if user.IsAdmin {
		return nil, apperr.NewForbiddenError("administrator jest zarządzany przez zmienne środowiskowe")
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
	return updated, nil
}

func (s *UserService) UpdateMe(ctx context.Context, id string, payload UpdateMePayload) (*User, *apperr.ServiceError) {
	user, svcErr := s.getUser(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}

	color, svcErr := normalizeColor(payload.Color)
	if svcErr != nil {
		return nil, svcErr
	}
	user.Color = color

	updated, err := s.dao.UpdateUser(ctx, user)
	if err != nil {
		return nil, daoError(err)
	}
	return updated, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) *apperr.ServiceError {
	user, svcErr := s.getUser(ctx, id)
	if svcErr != nil {
		return svcErr
	}
	if user.IsAdmin {
		return apperr.NewForbiddenError("nie można usunąć administratora")
	}
	if err := s.dao.DeleteUser(ctx, id); err != nil {
		return daoError(err)
	}

	log.FromContext(ctx).Info().Str("username", user.Username).Msg("User deleted")
	return nil
}

func (s *UserService) Login(ctx context.Context, payload LoginPayload) (*LoginResponse, *apperr.ServiceError) {
	username := strings.TrimSpace(payload.Username)
	limitKey := strings.ToLower(username)
	if !s.loginLimiter.Allowed(limitKey) {
		log.FromContext(ctx).Warn().Str("username", username).Msg("Login blocked by rate limit")
		return nil, &apperr.ServiceError{
			Code: http.StatusTooManyRequests,
			Err:  errors.New("za dużo nieudanych prób logowania, spróbuj za kilka minut"),
		}
	}

	user, err := s.dao.GetUserByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, daoError(err)
		}
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(payload.Password))
		return nil, s.loginFailed(ctx, username, limitKey)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)) != nil {
		return nil, s.loginFailed(ctx, username, limitKey)
	}
	s.loginLimiter.Reset(limitKey)

	token, err := auth.NewJwt(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, apperr.NewInternalError("failed to create token: %w", err)
	}

	return &LoginResponse{Token: token, User: *user}, nil
}

func (s *UserService) loginFailed(ctx context.Context, username, limitKey string) *apperr.ServiceError {
	s.loginLimiter.Fail(limitKey)
	log.FromContext(ctx).Warn().Str("username", username).Msg("Failed login attempt")
	return apperr.NewUnauthorizedError("nieprawidłowa nazwa użytkownika lub hasło")
}

func (s *UserService) IsAdmin(ctx context.Context, id string) (bool, error) {
	user, err := s.dao.GetUserByID(ctx, id)
	if errors.Is(err, db.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}

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
			Color:        randomColor(),
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

func (s *UserService) getUser(ctx context.Context, id string) (*User, *apperr.ServiceError) {
	if uuid.Validate(id) != nil {
		return nil, apperr.NewNotFoundError("nie znaleziono: użytkownik")
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
			"nazwa użytkownika musi mieć od %d do %d znaków",
			minUsernameLength,
			maxUsernameLength,
		)
	}
	return nil
}

func validatePassword(password string) *apperr.ServiceError {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return apperr.NewBadRequestError(
			"hasło musi mieć od %d do %d znaków",
			minPasswordLength,
			maxPasswordLength,
		)
	}
	return nil
}

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "user")
}
