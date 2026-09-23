package plans

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
)

type IPlanService interface {
	GetPlansByActivity(ctx context.Context, activityID string) ([]Plan, *apperr.ServiceError)
	CreatePlan(ctx context.Context, activityID, userID string, payload PlanPayload) (*Plan, *apperr.ServiceError)
	UpdatePlan(ctx context.Context, id, userID string, payload PlanUpdatePayload) (*Plan, *apperr.ServiceError)
	DeletePlan(ctx context.Context, id, userID string) *apperr.ServiceError
}

type PlanServiceDeps struct {
	PlanDAO     IPlanDAO
	ActivityDAO activities.IActivityDAO
	Clock       clock.Clock
}

type PlanService struct {
	dao         IPlanDAO
	activityDAO activities.IActivityDAO
	clock       clock.Clock
}

func NewPlanService(deps PlanServiceDeps) *PlanService {
	return &PlanService{dao: deps.PlanDAO, activityDAO: deps.ActivityDAO, clock: deps.Clock}
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
	if plan.StartsOn.Before(s.clock.Today()) {
		return nil, apperr.NewBadRequestError("plan nie może zaczynać się w przeszłości")
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
	payload PlanUpdatePayload,
) (*Plan, *apperr.ServiceError) {
	plan, svcErr := s.getOwnPlan(ctx, id, userID)
	if svcErr != nil {
		return nil, svcErr
	}

	if payload.Title != nil {
		title := strings.TrimSpace(*payload.Title)
		if title == "" {
			return nil, apperr.NewBadRequestError("tytuł jest wymagany")
		}
		plan.Title = title
	}

	if payload.EndsOn != nil {
		endsOn, svcErr := s.checkEndsOn(ctx, plan, *payload.EndsOn)
		if svcErr != nil {
			return nil, svcErr
		}
		plan.EndsOn = endsOn
	}

	updated, err := s.dao.UpdatePlan(ctx, plan)
	if err != nil {
		return nil, daoError(err)
	}
	return updated, nil
}

func (s *PlanService) checkEndsOn(ctx context.Context, plan *Plan, value string) (*clock.Date, *apperr.ServiceError) {
	if value == "" {
		return nil, nil
	}
	endsOn, err := clock.ParseDate(value)
	if err != nil {
		return nil, apperr.NewBadRequestError("endsOn musi być datą RRRR-MM-DD")
	}
	if endsOn.Before(plan.StartsOn) {
		return nil, apperr.NewBadRequestError("koniec nie może być przed startem; plan, który się nie zaczął, usuń")
	}
	if endsOn.Before(s.clock.Today().AddDays(-1)) {
		return nil, apperr.NewBadRequestError("plan można zakończyć najwcześniej wczoraj, miniona historia zostaje")
	}
	last, err := s.dao.GetLastScheduledFor(ctx, plan.ID)
	if err != nil {
		return nil, daoError(err)
	}
	if last != nil && endsOn.Before(*last) {
		return nil, apperr.NewBadRequestError("plan ma już wpis na %s", *last)
	}
	return &endsOn, nil
}

func (s *PlanService) DeletePlan(ctx context.Context, id, userID string) *apperr.ServiceError {
	plan, svcErr := s.getOwnPlan(ctx, id, userID)
	if svcErr != nil {
		return svcErr
	}
	if plan.StartsOn.Before(s.clock.Today()) {
		return apperr.NewConflictError("plan już się zaczął, zakończ go zamiast usuwać")
	}
	last, err := s.dao.GetLastScheduledFor(ctx, plan.ID)
	if err != nil {
		return daoError(err)
	}
	if last != nil {
		return apperr.NewConflictError("plan ma wpisy, zakończ go zamiast usuwać")
	}
	if err := s.dao.DeletePlan(ctx, id); err != nil {
		return daoError(err)
	}
	return nil
}

func (s *PlanService) getOwnPlan(ctx context.Context, id, userID string) (*Plan, *apperr.ServiceError) {
	plan, err := s.dao.GetPlanByID(ctx, id)
	if err != nil {
		return nil, daoError(err)
	}
	if plan.UserID != userID {
		return nil, apperr.NewForbiddenError("możesz zmieniać tylko własne plany")
	}
	return plan, nil
}

func applyPayload(plan *Plan, payload PlanPayload) *apperr.ServiceError {
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		return apperr.NewBadRequestError("tytuł jest wymagany")
	}

	anyDay := payload.Monday || payload.Tuesday || payload.Wednesday || payload.Thursday ||
		payload.Friday || payload.Saturday || payload.Sunday
	if !anyDay {
		return apperr.NewBadRequestError("wybierz co najmniej jeden dzień tygodnia")
	}

	if payload.StartsOn.IsZero() {
		return apperr.NewBadRequestError("data startu jest wymagana")
	}
	if payload.EndsOn != nil && payload.EndsOn.Before(payload.StartsOn) {
		return apperr.NewBadRequestError("koniec nie może być przed startem")
	}

	plan.Title = title
	plan.Monday = payload.Monday
	plan.Tuesday = payload.Tuesday
	plan.Wednesday = payload.Wednesday
	plan.Thursday = payload.Thursday
	plan.Friday = payload.Friday
	plan.Saturday = payload.Saturday
	plan.Sunday = payload.Sunday
	plan.StartsOn = payload.StartsOn
	plan.EndsOn = payload.EndsOn
	return nil
}

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "plan")
}
