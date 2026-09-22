package plans

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/platform/apperr"
)

const dateLayout = time.DateOnly

type IPlanService interface {
	GetPlansByActivity(ctx context.Context, activityID string) ([]Plan, *apperr.ServiceError)
	GetPlan(ctx context.Context, id string) (*Plan, *apperr.ServiceError)
	CreatePlan(ctx context.Context, activityID, userID string, payload PlanPayload) (*Plan, *apperr.ServiceError)
	UpdatePlan(ctx context.Context, id, userID string, payload PlanPayload) (*Plan, *apperr.ServiceError)
	DeletePlan(ctx context.Context, id, userID string) *apperr.ServiceError
}

type PlanServiceDeps struct {
	PlanDAO     IPlanDAO
	ActivityDAO activities.IActivityDAO
}

type PlanService struct {
	dao         IPlanDAO
	activityDAO activities.IActivityDAO
}

func NewPlanService(deps PlanServiceDeps) *PlanService {
	return &PlanService{dao: deps.PlanDAO, activityDAO: deps.ActivityDAO}
}

func (s *PlanService) GetPlansByActivity(ctx context.Context, activityID string) ([]Plan, *apperr.ServiceError) {
	if _, err := s.activityDAO.GetActivityByID(ctx, activityID); err != nil {
		return nil, apperr.FromDAO(err, "activity")
	}
	plans, err := s.dao.GetPlansByActivity(ctx, activityID)
	if err != nil {
		return nil, daoError(err)
	}
	return plans, nil
}

func (s *PlanService) GetPlan(ctx context.Context, id string) (*Plan, *apperr.ServiceError) {
	plan, err := s.dao.GetPlanByID(ctx, id)
	if err != nil {
		return nil, daoError(err)
	}
	return plan, nil
}

func (s *PlanService) CreatePlan(
	ctx context.Context,
	activityID, userID string,
	payload PlanPayload,
) (*Plan, *apperr.ServiceError) {
	if _, err := s.activityDAO.GetActivityByID(ctx, activityID); err != nil {
		return nil, apperr.FromDAO(err, "activity")
	}

	plan := &Plan{ID: uuid.NewString(), ActivityID: activityID, UserID: userID}
	if svcErr := applyPayload(plan, payload); svcErr != nil {
		return nil, svcErr
	}

	created, err := s.dao.InsertPlan(ctx, plan)
	if err != nil {
		return nil, daoError(err)
	}
	return created, nil
}

func (s *PlanService) UpdatePlan(
	ctx context.Context,
	id, userID string,
	payload PlanPayload,
) (*Plan, *apperr.ServiceError) {
	plan, svcErr := s.getOwnPlan(ctx, id, userID)
	if svcErr != nil {
		return nil, svcErr
	}
	if svcErr := applyPayload(plan, payload); svcErr != nil {
		return nil, svcErr
	}

	updated, err := s.dao.UpdatePlan(ctx, plan)
	if err != nil {
		return nil, daoError(err)
	}
	return updated, nil
}

func (s *PlanService) DeletePlan(ctx context.Context, id, userID string) *apperr.ServiceError {
	if _, svcErr := s.getOwnPlan(ctx, id, userID); svcErr != nil {
		return svcErr
	}
	if err := s.dao.DeletePlan(ctx, id); err != nil {
		return daoError(err)
	}
	return nil
}

// getOwnPlan loads a plan and checks that it belongs to the user.
func (s *PlanService) getOwnPlan(ctx context.Context, id, userID string) (*Plan, *apperr.ServiceError) {
	plan, svcErr := s.GetPlan(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}
	if plan.UserID != userID {
		return nil, apperr.NewForbiddenError("you can only change your own plans")
	}
	return plan, nil
}

// applyPayload validates the payload and copies it onto the plan.
func applyPayload(plan *Plan, payload PlanPayload) *apperr.ServiceError {
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		return apperr.NewBadRequestError("title is required")
	}

	anyDay := payload.Monday || payload.Tuesday || payload.Wednesday || payload.Thursday ||
		payload.Friday || payload.Saturday || payload.Sunday
	if !anyDay {
		return apperr.NewBadRequestError("pick at least one day of the week")
	}

	startsOn, err := time.Parse(dateLayout, payload.StartsOn)
	if err != nil {
		return apperr.NewBadRequestError("startsOn must be a YYYY-MM-DD date")
	}

	var endsOn *time.Time
	if payload.EndsOn != nil && *payload.EndsOn != "" {
		parsed, err := time.Parse(dateLayout, *payload.EndsOn)
		if err != nil {
			return apperr.NewBadRequestError("endsOn must be a YYYY-MM-DD date")
		}
		if parsed.Before(startsOn) {
			return apperr.NewBadRequestError("endsOn must not be before startsOn")
		}
		endsOn = &parsed
	}

	plan.Title = title
	plan.Monday = payload.Monday
	plan.Tuesday = payload.Tuesday
	plan.Wednesday = payload.Wednesday
	plan.Thursday = payload.Thursday
	plan.Friday = payload.Friday
	plan.Saturday = payload.Saturday
	plan.Sunday = payload.Sunday
	plan.StartsOn = startsOn
	plan.EndsOn = endsOn
	return nil
}

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "plan")
}
