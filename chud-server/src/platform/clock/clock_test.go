package clock

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useWarsaw(t *testing.T) {
	t.Helper()
	require.NoError(t, Init("Europe/Warsaw"))
	t.Cleanup(func() { location = time.UTC })
}

func TestDateOfAndStart(t *testing.T) {
	useWarsaw(t)

	lateUTC := time.Date(2026, 9, 22, 22, 30, 0, 0, time.UTC)
	assert.Equal(t, NewDate(2026, 9, 23), DateOf(lateUTC), "00:30 in Warsaw is the next day")

	start := NewDate(2026, 9, 23).Start()
	assert.Equal(t, time.Date(2026, 9, 22, 22, 0, 0, 0, time.UTC), start.UTC())
}

func TestInitRejectsUnknownZone(t *testing.T) {
	assert.Error(t, Init("Mars/Olympus"))
}

func TestClockToday(t *testing.T) {
	useWarsaw(t)

	c := Fixed(time.Date(2026, 9, 22, 22, 30, 0, 0, time.UTC))
	assert.Equal(t, NewDate(2026, 9, 23), c.Today())
	assert.WithinDuration(t, time.Now(), Clock{}.Now(), time.Second)
}

func TestDateArithmetic(t *testing.T) {
	d := NewDate(2026, 3, 28)
	assert.Equal(t, NewDate(2026, 3, 30), d.AddDays(2))
	assert.Equal(t, 2, d.DaysUntil(d.AddDays(2)))
	assert.True(t, d.Before(d.AddDays(1)))
	assert.True(t, d.AddDays(1).After(d))
	assert.Equal(t, time.Saturday, d.Weekday())
	assert.True(t, Date{}.IsZero())
}

func TestDateJSON(t *testing.T) {
	data, err := json.Marshal(struct {
		On  Date  `json:"on"`
		End *Date `json:"end"`
	}{On: NewDate(2026, 9, 3)})
	require.NoError(t, err)
	assert.JSONEq(t, `{"on":"2026-09-03","end":null}`, string(data))

	var parsed struct {
		On Date `json:"on"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"on":"2026-09-03"}`), &parsed))
	assert.Equal(t, NewDate(2026, 9, 3), parsed.On)

	assert.Error(t, json.Unmarshal([]byte(`{"on":"03.09.2026"}`), &parsed))
}

func TestDateScanAndValue(t *testing.T) {
	var d Date
	require.NoError(t, d.Scan(time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, NewDate(2026, 9, 3), d)

	require.NoError(t, d.Scan([]byte("2026-09-04")))
	assert.Equal(t, NewDate(2026, 9, 4), d)

	value, err := d.Value()
	require.NoError(t, err)
	assert.Equal(t, "2026-09-04", value)

	assert.Error(t, d.Scan(42))
}

func TestValidateRange(t *testing.T) {
	from := NewDate(2026, 1, 1)
	assert.NoError(t, ValidateRange(from, from))
	assert.NoError(t, ValidateRange(from, from.AddDays(MaxRangeDays-1)))
	assert.Error(t, ValidateRange(from, from.AddDays(MaxRangeDays)))
	assert.Error(t, ValidateRange(from, from.AddDays(-1)))
	assert.Error(t, ValidateRange(Date{}, from))
}
