package httpx

import (
	"net/http/httptest"
	"testing"

	"github.com/sebnow/chud/platform/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateRangeQuery(t *testing.T) {
	from, to, err := DateRangeQuery(httptest.NewRequest("GET", "/x?from=2026-09-01&to=2026-09-30", nil))
	require.NoError(t, err)
	assert.Equal(t, clock.NewDate(2026, 9, 1), from)
	assert.Equal(t, clock.NewDate(2026, 9, 30), to)

	_, _, err = DateRangeQuery(httptest.NewRequest("GET", "/x?from=2026-09-01T00:00:00Z&to=2026-09-30", nil))
	assert.EqualError(t, err, "from musi być datą RRRR-MM-DD")

	_, _, err = DateRangeQuery(httptest.NewRequest("GET", "/x?from=2026-09-01", nil))
	assert.EqualError(t, err, "to musi być datą RRRR-MM-DD")
}
