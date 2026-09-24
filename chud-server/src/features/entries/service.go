package entries

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/log"
)

const (
	MaxFileSize          = 10 << 20
	MaxFileCount         = 10
	MaxDescriptionLength = 2000
	maxClockSkew         = 5 * time.Minute
	sniffLength          = 512
)

const unsupportedMediaMessage = "można dołączać tylko zdjęcia (JPG, PNG, GIF, WebP) i filmy (MP4, WebM, MOV)"

var allowedMediaTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
}

func IsAllowedMediaType(contentType string) bool {
	return allowedMediaTypes[contentType]
}

func DetectMediaType(head []byte) string {
	if len(head) >= 12 && bytes.Equal(head[4:8], []byte("ftyp")) && bytes.Equal(head[8:12], []byte("qt  ")) {
		return "video/quicktime"
	}
	return http.DetectContentType(head)
}

type IEntryService interface {
	GetEntriesByActivity(ctx context.Context, activityID string) ([]Entry, *apperr.ServiceError)
	CreateEntry(ctx context.Context, activityID, userID string, payload CreateEntryPayload) (*Entry, *apperr.ServiceError)
	GetMediaByEntry(ctx context.Context, entryID string) ([]Media, *apperr.ServiceError)
	GetMedia(ctx context.Context, id string) (*Media, *apperr.ServiceError)
	GetEntriesInRange(ctx context.Context, from, to clock.Date) ([]Entry, *apperr.ServiceError)
	GetUserEntriesInRange(ctx context.Context, userID string, from, to clock.Date) ([]Entry, *apperr.ServiceError)
}

type EntryServiceDeps struct {
	DB          db.DB
	NewEntryDAO func(db.Querier) IEntryDAO
	ActivityDAO activities.IActivityDAO
	PlanDAO     plans.IPlanDAO
	Clock       clock.Clock
}

type EntryService struct {
	db          db.DB
	newEntryDAO func(db.Querier) IEntryDAO
	activityDAO activities.IActivityDAO
	planDAO     plans.IPlanDAO
	clock       clock.Clock
}

func NewEntryService(deps EntryServiceDeps) *EntryService {
	newEntryDAO := deps.NewEntryDAO
	if newEntryDAO == nil {
		newEntryDAO = NewEntryDAO
	}
	return &EntryService{
		db:          deps.DB,
		newEntryDAO: newEntryDAO,
		activityDAO: deps.ActivityDAO,
		planDAO:     deps.PlanDAO,
		clock:       deps.Clock,
	}
}

func (s *EntryService) dao() IEntryDAO {
	return s.newEntryDAO(s.db.Querier())
}

func (s *EntryService) GetEntriesByActivity(ctx context.Context, activityID string) ([]Entry, *apperr.ServiceError) {
	if _, err := s.activityDAO.GetActivityByID(ctx, activityID); err != nil {
		return nil, apperr.FromDAO(err, "activity")
	}
	entries, err := s.dao().GetEntriesByActivity(ctx, activityID)
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
	activity, err := s.activityDAO.GetActivityByID(ctx, activityID)
	if err != nil {
		return nil, apperr.FromDAO(err, "activity")
	}
	if activity.IsArchived() {
		return nil, activities.ErrArchived()
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
	now := s.clock.Now()
	if entry.OccurredAt.IsZero() {
		entry.OccurredAt = now
	}
	if entry.OccurredAt.After(now.Add(maxClockSkew)) {
		return nil, apperr.NewBadRequestError("wpis nie może być z przyszłości")
	}

	if entry.PlanID != nil {
		if svcErr := s.checkPlannedEntry(ctx, entry); svcErr != nil {
			return nil, svcErr
		}
	} else if entry.ScheduledFor != nil || entry.Excused {
		return nil, apperr.NewBadRequestError("tylko wpis do planu może mieć zaplanowany dzień albo być wymówką")
	}

	if utf8.RuneCountInString(entry.Description) > MaxDescriptionLength {
		return nil, apperr.NewBadRequestError("opis może mieć maksymalnie %d znaków", MaxDescriptionLength)
	}

	if entry.Excused && entry.Description == "" {
		return nil, apperr.NewBadRequestError("wymówka musi mieć opis")
	}

	if entry.PlanID != nil && !entry.Excused && clock.DateOf(entry.OccurredAt) != *entry.ScheduledFor {
		return nil, apperr.NewBadRequestError("zaplanowany dzień trzeba zrobić tego samego dnia; dodaj zwykły wpis poza planem")
	}

	if len(payload.Files) > MaxFileCount {
		return nil, apperr.NewBadRequestError("maksymalnie %d plików na wpis", MaxFileCount)
	}
	media := make([]Media, 0, len(payload.Files))
	for _, f := range payload.Files {
		if len(f.Data) > MaxFileSize {
			return nil, apperr.NewBadRequestError("pliki mogą mieć maksymalnie 10 MB")
		}
		if !IsAllowedMediaType(f.ContentType) {
			return nil, apperr.NewBadRequestError(unsupportedMediaMessage)
		}
		media = append(media, Media{ID: uuid.NewString(), ContentType: f.ContentType, Data: f.Data})
	}

	var created *Entry
	err = s.db.WithTx(ctx, func(q db.Querier) error {
		dao := s.newEntryDAO(q)
		var err error
		if created, err = dao.InsertEntry(ctx, entry); err != nil {
			return err
		}
		return dao.InsertMedia(ctx, created.ID, media)
	})
	if errors.Is(err, db.ErrAlreadyExists) {
		return nil, apperr.NewConflictError("ten zaplanowany dzień jest już rozliczony")
	}
	if err != nil {
		return nil, daoError(err)
	}

	log.FromContext(ctx).Info().Str("activity_id", activityID).Int("media", len(media)).Msg("Entry created")
	return created, nil
}

func (s *EntryService) checkPlannedEntry(ctx context.Context, entry *Entry) *apperr.ServiceError {
	plan, err := s.planDAO.GetPlanByID(ctx, *entry.PlanID)
	if err != nil {
		return apperr.FromDAO(err, "plan")
	}
	if plan.ActivityID != entry.ActivityID {
		return apperr.NewBadRequestError("plan należy do innej aktywności")
	}
	if plan.UserID != entry.UserID {
		return apperr.NewForbiddenError("możesz rozliczać tylko własne plany")
	}
	if entry.ScheduledFor == nil {
		return apperr.NewBadRequestError("wpis do planu musi mieć zaplanowany dzień")
	}
	if !plan.IsScheduledOn(*entry.ScheduledFor) {
		return apperr.NewBadRequestError("plan nie ma tego dnia w harmonogramie")
	}
	return nil
}

func (s *EntryService) GetMediaByEntry(ctx context.Context, entryID string) ([]Media, *apperr.ServiceError) {
	if _, err := s.dao().GetEntryByID(ctx, entryID); err != nil {
		return nil, daoError(err)
	}
	media, err := s.dao().GetMediaByEntry(ctx, entryID)
	if err != nil {
		return nil, apperr.FromDAO(err, "media")
	}
	return media, nil
}

func (s *EntryService) GetMedia(ctx context.Context, id string) (*Media, *apperr.ServiceError) {
	media, err := s.dao().GetMediaByID(ctx, id)
	if err != nil {
		return nil, apperr.FromDAO(err, "media")
	}
	return media, nil
}

func (s *EntryService) GetEntriesInRange(ctx context.Context, from, to clock.Date) ([]Entry, *apperr.ServiceError) {
	if err := clock.ValidateRange(from, to); err != nil {
		return nil, apperr.NewBadRequestError("%w", err)
	}
	entries, err := s.dao().GetEntriesInRange(ctx, from.Start(), to.AddDays(1).Start())
	if err != nil {
		return nil, daoError(err)
	}
	return entries, nil
}

func (s *EntryService) GetUserEntriesInRange(
	ctx context.Context,
	userID string,
	from, to clock.Date,
) ([]Entry, *apperr.ServiceError) {
	if err := clock.ValidateRange(from, to); err != nil {
		return nil, apperr.NewBadRequestError("%w", err)
	}
	entries, err := s.dao().GetUserEntriesInRange(ctx, userID, from.Start(), to.AddDays(1).Start())
	if err != nil {
		return nil, daoError(err)
	}
	return entries, nil
}

func daoError(err error) *apperr.ServiceError {
	return apperr.FromDAO(err, "entry")
}
