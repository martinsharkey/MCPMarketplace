package gateway

import (
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter keyed by principal or service.
type Limiter struct {
	Rate  float64 // tokens added per second
	Burst int     // bucket size
	Now   func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

// Allow takes one token for key, reporting false when the bucket is empty.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if l.Now != nil {
		now = l.Now()
	}
	if l.buckets == nil {
		l.buckets = map[string]*bucket{}
	}
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(l.Burst), last: now}
		l.buckets[key] = b
	}
	b.tokens = min(float64(l.Burst), b.tokens+now.Sub(b.last).Seconds()*l.Rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Keys returns how many keys the limiter is tracking.
func (l *Limiter) Keys() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
