package apperr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sebnow/chud/platform/db"
)

type ServiceError struct {
	Code int
	Err  error
}

func NewNotFoundError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusNotFound,
		Err:  fmt.Errorf(format, args...),
	}
}

func NewInternalError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusInternalServerError,
		Err:  fmt.Errorf(format, args...),
	}
}

func NewBadRequestError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusBadRequest,
		Err:  fmt.Errorf(format, args...),
	}
}

func NewUnauthorizedError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusUnauthorized,
		Err:  fmt.Errorf(format, args...),
	}
}

func NewConflictError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusConflict,
		Err:  fmt.Errorf(format, args...),
	}
}

func (e ServiceError) Error() string {
	return e.Err.Error()
}

func NewForbiddenError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusForbidden,
		Err:  fmt.Errorf(format, args...),
	}
}

func FromDAO(err error, resource string) *ServiceError {
	switch {
	case errors.Is(err, db.ErrNotFound):
		return NewNotFoundError("nie znaleziono: %s", resourceName(resource))
	case errors.Is(err, db.ErrAlreadyExists):
		return NewConflictError("już istnieje: %s", resourceName(resource))
	case errors.Is(err, db.ErrInUse):
		return NewConflictError("wciąż w użyciu: %s", resourceName(resource))
	default:
		return NewInternalError("%w", err)
	}
}

var resourceNames = map[string]string{
	"activity": "aktywność",
	"entry":    "wpis",
	"media":    "plik",
	"plan":     "plan",
	"user":     "użytkownik",
}

func resourceName(resource string) string {
	if name, ok := resourceNames[resource]; ok {
		return name
	}
	return resource
}

func NewTooManyRequestsError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusTooManyRequests,
		Err:  fmt.Errorf(format, args...),
	}
}

func NewUnavailableError(format string, args ...any) *ServiceError {
	return &ServiceError{
		Code: http.StatusServiceUnavailable,
		Err:  fmt.Errorf(format, args...),
	}
}
