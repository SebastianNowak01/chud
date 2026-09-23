package activities

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/apperr"
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
}

type ActivityServiceDeps struct {
	ActivityDAO IActivityDAO
}

type ActivityService struct {
	dao IActivityDAO
}

func NewActivityService(deps ActivityServiceDeps) *ActivityService {
	return &ActivityService{dao: deps.ActivityDAO}
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

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "activity")
}
