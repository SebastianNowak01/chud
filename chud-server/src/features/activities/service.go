package activities

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/log"
)

const (
	maxNameLength        = 50
	maxDescriptionLength = 500
)

type IActivityService interface {
	GetAllActivities(ctx context.Context) ([]Activity, *apperr.ServiceError)
	GetActivity(ctx context.Context, id string) (*Activity, *apperr.ServiceError)
	CreateActivity(ctx context.Context, userID string, payload CreateActivityPayload) (*Activity, *apperr.ServiceError)
	GetMembers(ctx context.Context, activityID string) ([]users.User, *apperr.ServiceError)
	DeleteActivity(ctx context.Context, id string, actor Actor) *apperr.ServiceError
	ArchiveActivity(ctx context.Context, id string, actor Actor) (*Activity, *apperr.ServiceError)
	RestoreActivity(ctx context.Context, id string, actor Actor) (*Activity, *apperr.ServiceError)
}

type Actor struct {
	UserID  string
	IsAdmin bool
}

type ActivityServiceDeps struct {
	ActivityDAO IActivityDAO
	Clock       clock.Clock
}

type ActivityService struct {
	dao   IActivityDAO
	clock clock.Clock
}

func NewActivityService(deps ActivityServiceDeps) *ActivityService {
	return &ActivityService{dao: deps.ActivityDAO, clock: deps.Clock}
}

func ErrArchived() *apperr.ServiceError {
	return apperr.NewConflictError("aktywność jest w archiwum, przywróć ją, żeby coś dodać")
}

func (s *ActivityService) GetAllActivities(ctx context.Context) ([]Activity, *apperr.ServiceError) {
	activities, err := s.dao.GetAllActivities(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	return activities, nil
}

func (s *ActivityService) GetActivity(ctx context.Context, id string) (*Activity, *apperr.ServiceError) {
	activity, err := s.dao.GetActivityByID(ctx, id)
	if err != nil {
		return nil, daoError(err)
	}
	return activity, nil
}

func (s *ActivityService) CreateActivity(
	ctx context.Context,
	userID string,
	payload CreateActivityPayload,
) (*Activity, *apperr.ServiceError) {
	name := strings.TrimSpace(payload.Name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLength {
		return nil, apperr.NewBadRequestError("nazwa musi mieć od 1 do %d znaków", maxNameLength)
	}
	description := strings.TrimSpace(payload.Description)
	if utf8.RuneCountInString(description) > maxDescriptionLength {
		return nil, apperr.NewBadRequestError("opis może mieć maksymalnie %d znaków", maxDescriptionLength)
	}

	activity, err := s.dao.InsertActivity(ctx, &Activity{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		CreatedBy:   userID,
	})
	if errors.Is(err, db.ErrAlreadyExists) {
		if existing, lookupErr := s.dao.GetActivityByName(ctx, name); lookupErr == nil && existing.IsArchived() {
			return nil, apperr.NewConflictError("taka aktywność jest w archiwum, przywróć ją zamiast tworzyć nową")
		}
		return nil, apperr.NewConflictError("aktywność o tej nazwie już istnieje")
	}
	if err != nil {
		return nil, daoError(err)
	}

	log.FromContext(ctx).Info().Str("name", name).Msg("Activity created")
	return activity, nil
}

func (s *ActivityService) GetMembers(ctx context.Context, activityID string) ([]users.User, *apperr.ServiceError) {
	if _, svcErr := s.GetActivity(ctx, activityID); svcErr != nil {
		return nil, svcErr
	}
	members, err := s.dao.GetMembers(ctx, activityID)
	if err != nil {
		return nil, daoError(err)
	}
	return members, nil
}

func (s *ActivityService) DeleteActivity(ctx context.Context, id string, actor Actor) *apperr.ServiceError {
	activity, svcErr := s.getManaged(ctx, id, actor)
	if svcErr != nil {
		return svcErr
	}
	if activity.HasHistory {
		return apperr.NewConflictError("aktywność ma już plany albo wpisy, możesz ją tylko zarchiwizować")
	}
	err := s.dao.DeleteEmptyActivity(ctx, id)
	if errors.Is(err, db.ErrNotFound) {
		return apperr.NewConflictError("aktywność ma już plany albo wpisy, możesz ją tylko zarchiwizować")
	}
	if err != nil {
		return daoError(err)
	}
	log.FromContext(ctx).Info().Str("activity_id", id).Msg("Activity deleted")
	return nil
}

func (s *ActivityService) ArchiveActivity(ctx context.Context, id string, actor Actor) (*Activity, *apperr.ServiceError) {
	activity, svcErr := s.getManaged(ctx, id, actor)
	if svcErr != nil {
		return nil, svcErr
	}
	if activity.IsArchived() {
		return nil, apperr.NewConflictError("aktywność już jest w archiwum")
	}
	archived, err := s.dao.ArchiveActivity(ctx, id, s.clock.Today())
	if errors.Is(err, db.ErrNotFound) {
		return nil, apperr.NewConflictError("aktywność już jest w archiwum")
	}
	if err != nil {
		return nil, daoError(err)
	}
	log.FromContext(ctx).Info().Str("activity_id", id).Msg("Activity archived")
	return archived, nil
}

func (s *ActivityService) RestoreActivity(ctx context.Context, id string, actor Actor) (*Activity, *apperr.ServiceError) {
	activity, svcErr := s.getManaged(ctx, id, actor)
	if svcErr != nil {
		return nil, svcErr
	}
	if !activity.IsArchived() {
		return nil, apperr.NewConflictError("aktywność nie jest w archiwum")
	}
	restored, err := s.dao.RestoreActivity(ctx, id)
	if err != nil {
		return nil, daoError(err)
	}
	log.FromContext(ctx).Info().Str("activity_id", id).Msg("Activity restored")
	return restored, nil
}

func (s *ActivityService) getManaged(ctx context.Context, id string, actor Actor) (*Activity, *apperr.ServiceError) {
	activity, err := s.dao.GetActivityByID(ctx, id)
	if err != nil {
		return nil, daoError(err)
	}
	if activity.CreatedBy != actor.UserID && !actor.IsAdmin {
		return nil, apperr.NewForbiddenError("tylko autor aktywności albo admin może ją usunąć lub zarchiwizować")
	}
	return activity, nil
}

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "activity")
}
