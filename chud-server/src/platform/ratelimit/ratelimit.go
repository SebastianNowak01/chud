package ratelimit

import (
	"sync"
	"time"
)

const pruneThreshold = 10000

type bucket struct {
	count int
	reset time.Time
}

type Limiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	now     func() time.Time
	buckets map[string]*bucket
}

func New(maxFailures int, window time.Duration) *Limiter {
	return &Limiter{max: maxFailures, window: window, now: time.Now, buckets: map[string]*bucket{}}
}

func (l *Limiter) Allowed(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	return !ok || l.now().After(b.reset) || b.count < l.max
}

func (l *Limiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.buckets) >= pruneThreshold {
		for k, b := range l.buckets {
			if now.After(b.reset) {
				delete(l.buckets, k)
			}
		}
	}
	b, ok := l.buckets[key]
	if !ok || now.After(b.reset) {
		b = &bucket{reset: now.Add(l.window)}
		l.buckets[key] = b
	}
	b.count++
}

func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}
