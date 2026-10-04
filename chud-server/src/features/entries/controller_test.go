package entries

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sebnow/chud/platform/auth"
	"github.com/stretchr/testify/assert"
)

func TestCreateEntryWaitsForUploadSlot(t *testing.T) {
	c := NewEntryAPIController(nil)
	for range maxUploads {
		c.uploads <- struct{}{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	ctx = auth.SetUserInContext(ctx, &auth.UserClaims{UserID: "u1", Username: "alice"})
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/activities/x/entries", nil)
	req.SetPathValue("id", "8f3c2a8e-8a4c-4d55-9a8a-1f5b8f0f6b1e")
	rec := httptest.NewRecorder()

	c.CreateEntryHandler(rec, req)

	assert.Equal(t, maxUploads, len(c.uploads), "slots untouched while full")
	assert.Empty(t, rec.Body.String(), "request gave up without reaching the service")
}
