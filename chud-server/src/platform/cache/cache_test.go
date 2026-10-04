package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestCache(maxEntries int) (*LRU[string, int], *time.Time) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	c := New[string, int](maxEntries)
	c.now = func() time.Time { return now }
	return c, &now
}

func TestLRU(t *testing.T) {
	t.Run("get and overwrite", func(t *testing.T) {
		c, _ := newTestCache(3)
		_, ok := c.Get("a")
		assert.False(t, ok)

		c.Set("a", 1, time.Hour)
		c.Set("a", 2, time.Hour)
		v, ok := c.Get("a")
		assert.True(t, ok)
		assert.Equal(t, 2, v)
		assert.Equal(t, 1, c.Len())
	})

	t.Run("expires after ttl", func(t *testing.T) {
		c, now := newTestCache(3)
		c.Set("a", 1, time.Hour)

		*now = now.Add(time.Hour - time.Second)
		_, ok := c.Get("a")
		assert.True(t, ok)

		*now = now.Add(time.Second)
		_, ok = c.Get("a")
		assert.False(t, ok)
		assert.Equal(t, 0, c.Len(), "expired entry is dropped")
	})

	t.Run("overwrite renews ttl", func(t *testing.T) {
		c, now := newTestCache(3)
		c.Set("a", 1, time.Hour)
		*now = now.Add(30 * time.Minute)
		c.Set("a", 1, time.Hour)
		*now = now.Add(45 * time.Minute)
		_, ok := c.Get("a")
		assert.True(t, ok)
	})

	t.Run("evicts least recently used", func(t *testing.T) {
		c, _ := newTestCache(2)
		c.Set("a", 1, time.Hour)
		c.Set("b", 2, time.Hour)
		c.Get("a")
		c.Set("c", 3, time.Hour)

		_, ok := c.Get("b")
		assert.False(t, ok)
		_, ok = c.Get("a")
		assert.True(t, ok)
		_, ok = c.Get("c")
		assert.True(t, ok)
		assert.Equal(t, 2, c.Len())
	})

	t.Run("delete", func(t *testing.T) {
		c, _ := newTestCache(2)
		c.Set("a", 1, time.Hour)
		c.Delete("a")
		c.Delete("missing")
		_, ok := c.Get("a")
		assert.False(t, ok)
	})
}
