package entries

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/log"
)

const (
	MaxFileSize  = 10 << 20 // 10 MB
	MaxFileCount = 10
	maxRange     = 400 * 24 * time.Hour // a year of whole weeks, with slack
)

type IEntryService interface {
	GetEntriesByActivity(ctx context.Context, activityID string) ([]Entry, *apperr.ServiceError)
	CreateEntry(ctx context.Context, activityID, userID string, payload CreateEntryPayload) (*Entry, *apperr.ServiceError)
	GetMediaByEntry(ctx context.Context, entryID string) ([]Media, *apperr.ServiceError)
	GetMedia(ctx context.Context, id string) (*Media, *apperr.ServiceError)
	GetEntriesInRange(ctx context.Context, from, to time.Time) ([]Entry, *apperr.ServiceError)
	GetUserEntriesInRange(ctx context.Context, userID string, from, to time.Time) ([]Entry, *apperr.ServiceError)
}

type EntryServiceDeps struct {
	EntryDAO    IEntryDAO
	ActivityDAO activities.IActivityDAO
	PlanDAO     plans.IPlanDAO
}

type EntryService struct {
	dao         IEntryDAO
	activityDAO activities.IActivityDAO
	planDAO     plans.IPlanDAO
}

func NewEntryService(deps EntryServiceDeps) *EntryService {
	return &EntryService{dao: deps.EntryDAO, activityDAO: deps.ActivityDAO, planDAO: deps.PlanDAO}
}

func (s *EntryService) GetEntriesByActivity(ctx context.Context, activityID string) ([]Entry, *apperr.ServiceError) {
	if _, err := s.activityDAO.GetActivityByID(ctx, activityID); err != nil {
		return nil, apperr.FromDAO(err, "activity")
	}
	entries, err := s.dao.GetEntriesByActivity(ctx, activityID)
	if err != nil {
		return nil, daoError(err)
	}
	return entries, nil
}

func (s *EntryService) CreateEntry(
	ctx context.Context,
	activityID, userID string,
	payload CreateEntryPayload,
) (*Entry, *apperr.ServiceError) {
	if _, err := s.activityDAO.GetActivityByID(ctx, activityID); err != nil {
		return nil, apperr.FromDAO(err, "activity")
	}

	entry := &Entry{
		ID:           uuid.NewString(),
		ActivityID:   activityID,
		UserID:       userID,
		PlanID:       payload.PlanID,
		ScheduledFor: payload.ScheduledFor,
		Excused:      payload.Excused,
		Description:  strings.TrimSpace(payload.Description),
		OccurredAt:   payload.OccurredAt,
	}
	if entry.OccurredAt.IsZero() {
		entry.OccurredAt = time.Now()
	}

	if entry.PlanID != nil {
		if svcErr := s.checkPlannedEntry(ctx, entry); svcErr != nil {
			return nil, svcErr
		}
	} else if entry.ScheduledFor != nil || entry.Excused {
		return nil, apperr.NewBadRequestError("only planned entries can have a scheduled day or be excused")
	}

	if entry.Excused && entry.Description == "" {
		return nil, apperr.NewBadRequestError("an excuse needs a description")
	}

	if len(payload.Files) > MaxFileCount {
		return nil, apperr.NewBadRequestError("at most %d files per entry", MaxFileCount)
	}
	media := make([]Media, 0, len(payload.Files))
	for _, f := range payload.Files {
		if len(f.Data) > MaxFileSize {
			return nil, apperr.NewBadRequestError("files must be at most 10 MB")
		}
		if !strings.HasPrefix(f.ContentType, "image/") && !strings.HasPrefix(f.ContentType, "video/") {
			return nil, apperr.NewBadRequestError("only images and videos can be attached")
		}
		media = append(media, Media{ID: uuid.NewString(), ContentType: f.ContentType, Data: f.Data})
	}

	created, err := s.dao.InsertEntry(ctx, entry, media)
	if errors.Is(err, db.ErrAlreadyExists) {
		return nil, apperr.NewConflictError("this planned day already has an entry")
	}
	if err != nil {
		return nil, daoError(err)
	}

	log.FromContext(ctx).Info().Str("activity_id", activityID).Int("media", len(media)).Msg("Entry created")
	return created, nil
}

// checkPlannedEntry verifies the entry resolves a planned day of the user's own plan in this activity.
func (s *EntryService) checkPlannedEntry(ctx context.Context, entry *Entry) *apperr.ServiceError {
	plan, err := s.planDAO.GetPlanByID(ctx, *entry.PlanID)
	if err != nil {
		return apperr.FromDAO(err, "plan")
	}
	if plan.ActivityID != entry.ActivityID {
		return apperr.NewBadRequestError("the plan belongs to a different activity")
	}
	if plan.UserID != entry.UserID {
		return apperr.NewForbiddenError("you can only resolve your own plans")
	}
	if entry.ScheduledFor == nil {
		return apperr.NewBadRequestError("a planned entry needs the scheduled day")
	}
	if !plan.IsScheduledOn(*entry.ScheduledFor) {
		return apperr.NewBadRequestError("the plan has no planned day on that date")
	}
	return nil
}

func (s *EntryService) GetMediaByEntry(ctx context.Context, entryID string) ([]Media, *apperr.ServiceError) {
	if _, err := s.dao.GetEntryByID(ctx, entryID); err != nil {
		return nil, daoError(err)
	}
	media, err := s.dao.GetMediaByEntry(ctx, entryID)
	if err != nil {
		return nil, apperr.FromDAO(err, "media")
	}
	return media, nil
}

func (s *EntryService) GetMedia(ctx context.Context, id string) (*Media, *apperr.ServiceError) {
	media, err := s.dao.GetMediaByID(ctx, id)
	if err != nil {
		return nil, apperr.FromDAO(err, "media")
	}
	return media, nil
}

func (s *EntryService) GetEntriesInRange(ctx context.Context, from, to time.Time) ([]Entry, *apperr.ServiceError) {
	if svcErr := validateRange(from, to); svcErr != nil {
		return nil, svcErr
	}
	entries, err := s.dao.GetEntriesInRange(ctx, from, to)
	if err != nil {
		return nil, daoError(err)
	}
	return entries, nil
}

func (s *EntryService) GetUserEntriesInRange(
	ctx context.Context,
	userID string,
	from, to time.Time,
) ([]Entry, *apperr.ServiceError) {
	if svcErr := validateRange(from, to); svcErr != nil {
		return nil, svcErr
	}
	entries, err := s.dao.GetUserEntriesInRange(ctx, userID, from, to)
	if err != nil {
		return nil, daoError(err)
	}
	return entries, nil
}

func validateRange(from, to time.Time) *apperr.ServiceError {
	if !from.Before(to) {
		return apperr.NewBadRequestError("from must be before to")
	}
	if to.Sub(from) > maxRange {
		return apperr.NewBadRequestError("the range can be at most 400 days")
	}
	return nil
}

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "entry")
}
