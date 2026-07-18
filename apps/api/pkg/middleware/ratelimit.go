package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a fixed-window per-IP request limiter. In-memory, so it
// applies per instance; good enough to stop credential brute-forcing on a
// single-node deployment.
type RateLimiter struct {
	mu     sync.Mutex
	visits map[string]*window
	limit  int
	period time.Duration
}

type window struct {
	count int
	start time.Time
}

func NewRateLimiter(limit int, period time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visits: make(map[string]*window),
		limit:  limit,
		period: period,
	}
	// Prune stale windows so the map doesn't grow unbounded.
	go func() {
		for range time.Tick(period) {
			rl.mu.Lock()
			for ip, w := range rl.visits {
				if time.Since(w.start) > rl.period {
					delete(rl.visits, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	w, ok := rl.visits[key]
	if !ok || time.Since(w.start) > rl.period {
		rl.visits[key] = &window{count: 1, start: time.Now()}
		return true
	}
	w.count++
	return w.count <= rl.limit
}

// Middleware rejects requests over the limit with 429 Too Many Requests.
// Keyed by client IP; if this ever runs behind a reverse proxy, switch the
// key to a trusted X-Forwarded-For value.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !rl.allow(ip) {
			http.Error(w, "Too many attempts, please try again later", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
