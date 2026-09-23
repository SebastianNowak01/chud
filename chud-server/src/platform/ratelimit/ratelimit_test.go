package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	l := New(3, time.Minute)
	l.now = func() time.Time { return now }

	for range 3 {
		assert.True(t, l.Allowed("alice"))
		l.Fail("alice")
	}
	assert.False(t, l.Allowed("alice"))
	assert.True(t, l.Allowed("bob"), "keys are independent")

	now = now.Add(time.Minute + time.Second)
	assert.True(t, l.Allowed("alice"), "window expired")

	l.Fail("alice")
	l.Reset("alice")
	assert.True(t, l.Allowed("alice"))
}
