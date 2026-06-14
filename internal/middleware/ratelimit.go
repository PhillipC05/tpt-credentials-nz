package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu     sync.Mutex
	counts map[string][]time.Time
	limit  int
	window time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		counts: make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	hits := rl.counts[key]
	valid := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.counts[key] = valid

	if len(valid) >= rl.limit {
		return false
	}
	rl.counts[key] = append(rl.counts[key], now)
	return true
}

func (rl *rateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for k, hits := range rl.counts {
			valid := hits[:0]
			for _, t := range hits {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.counts, k)
			} else {
				rl.counts[k] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a middleware that limits requests per IP to limit per window.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = fwd
			}
			if !rl.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"too many requests"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
