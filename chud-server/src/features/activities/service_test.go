package activities

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/apperr"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeActivityDAO struct {
	activities map[string]Activity
	archivedOn *clock.Date
}

func (d *fakeActivityDAO) GetAllActivities(context.Context) ([]Activity, error) { return nil, nil }

func (d *fakeActivityDAO) GetActivityByID(_ context.Context, id string) (*Activity, error) {
	a, ok := d.activities[id]
	if !ok {
		return nil, db.ErrNotFound
	}
	return &a, nil
}

func (d *fakeActivityDAO) GetActivityByName(_ context.Context, name string) (*Activity, error) {
	for _, a := range d.activities {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, db.ErrNotFound
}

func (d *fakeActivityDAO) InsertActivity(_ context.Context, activity *Activity) (*Activity, error) {
	for _, a := range d.activities {
		if a.Name == activity.Name {
			return nil, db.ErrAlreadyExists
		}
	}
	d.activities[activity.ID] = *activity
	return activity, nil
}

func (d *fakeActivityDAO) GetMembers(context.Context, string) ([]users.User, error) { return nil, nil }

func (d *fakeActivityDAO) DeleteEmptyActivity(_ context.Context, id string) error {
	delete(d.activities, id)
	return nil
}

func (d *fakeActivityDAO) ArchiveActivity(_ context.Context, id string, today clock.Date) (*Activity, error) {
	a := d.activities[id]
	now := time.Now()
	a.ArchivedAt = &now
	d.activities[id] = a
	d.archivedOn = &today
	return &a, nil
}

func (d *fakeActivityDAO) RestoreActivity(_ context.Context, id string) (*Activity, error) {
	a := d.activities[id]
	a.ArchivedAt = nil
	d.activities[id] = a
	return &a, nil
}

func newTestService() (*ActivityService, *fakeActivityDAO) {
	archivedAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	dao := &fakeActivityDAO{activities: map[string]Activity{
		"empty": {ID: "empty", Name: "Pusta", CreatedBy: "alice"},
		"used":  {ID: "used", Name: "Siłownia", CreatedBy: "alice", HasHistory: true},
		"old":   {ID: "old", Name: "Stara", CreatedBy: "alice", HasHistory: true, ArchivedAt: &archivedAt},
	}}
	today, _ := clock.ParseDate("2026-09-24")
	return NewActivityService(ActivityServiceDeps{
		ActivityDAO: dao,
		Clock:       clock.Fixed(today.Start().Add(12 * time.Hour)),
	}), dao
}

var (
	author = Actor{UserID: "alice"}
	admin  = Actor{UserID: "root", IsAdmin: true}
	other  = Actor{UserID: "bob"}
)

func TestDeleteActivity(t *testing.T) {
	ctx := context.Background()

	t.Run("author deletes an empty activity", func(t *testing.T) {
		svc, dao := newTestService()
		require.Nil(t, svc.DeleteActivity(ctx, "empty", author))
		assert.NotContains(t, dao.activities, "empty")
	})

	t.Run("admin deletes someone else's empty activity", func(t *testing.T) {
		svc, _ := newTestService()
		assert.Nil(t, svc.DeleteActivity(ctx, "empty", admin))
	})

	tests := []struct {
		name     string
		id       string
		actor    Actor
		wantCode int
	}{
		{"someone else cannot", "empty", other, http.StatusForbidden},
		{"activity with history", "used", author, http.StatusConflict},
		{"unknown activity", "nope", author, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService()
			svcErr := svc.DeleteActivity(ctx, tt.id, tt.actor)
			require.NotNil(t, svcErr)
			assert.Equal(t, tt.wantCode, svcErr.Code)
		})
	}
}

func TestArchiveAndRestoreActivity(t *testing.T) {
	ctx := context.Background()

	t.Run("archives as of today and restores", func(t *testing.T) {
		svc, dao := newTestService()
		archived, svcErr := svc.ArchiveActivity(ctx, "used", author)
		require.Nil(t, svcErr)
		assert.True(t, archived.IsArchived())
		assert.Equal(t, "2026-09-24", dao.archivedOn.String())

		restored, svcErr := svc.RestoreActivity(ctx, "used", admin)
		require.Nil(t, svcErr)
		assert.False(t, restored.IsArchived())
	})

	tests := []struct {
		name     string
		action   func(*ActivityService) *apperr.ServiceError
		wantCode int
	}{
		{"someone else cannot archive", func(s *ActivityService) *apperr.ServiceError {
			_, err := s.ArchiveActivity(ctx, "used", other)
			return err
		}, http.StatusForbidden},
		{"already archived", func(s *ActivityService) *apperr.ServiceError {
			_, err := s.ArchiveActivity(ctx, "old", author)
			return err
		}, http.StatusConflict},
		{"restore an active one", func(s *ActivityService) *apperr.ServiceError {
			_, err := s.RestoreActivity(ctx, "used", author)
			return err
		}, http.StatusConflict},
		{"someone else cannot restore", func(s *ActivityService) *apperr.ServiceError {
			_, err := s.RestoreActivity(ctx, "old", other)
			return err
		}, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService()
			svcErr := tt.action(svc)
			require.NotNil(t, svcErr)
			assert.Equal(t, tt.wantCode, svcErr.Code)
		})
	}
}

func TestCreateActivityNameTakenByArchived(t *testing.T) {
	svc, _ := newTestService()

	_, svcErr := svc.CreateActivity(context.Background(), "bob", CreateActivityPayload{Name: "Stara"})
	require.NotNil(t, svcErr)
	assert.Equal(t, http.StatusConflict, svcErr.Code)
	assert.Contains(t, svcErr.Err.Error(), "archiwum")

	_, svcErr = svc.CreateActivity(context.Background(), "bob", CreateActivityPayload{Name: "Siłownia"})
	require.NotNil(t, svcErr)
	assert.NotContains(t, svcErr.Err.Error(), "archiwum")
}
