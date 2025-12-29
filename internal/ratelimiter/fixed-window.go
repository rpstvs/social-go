package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (r *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	r.RLock()
	count, exists := r.clients[ip]
	r.Unlock()

	if !exists || count < r.limit {
		r.Lock()

		if !exists {
			go r.resetCount(ip)
		}

		r.clients[ip]++
		r.Unlock()
		return true, 0
	}

	return false, r.window
}

func (r *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(r.window)
	r.Lock()
	delete(r.clients, ip)
	r.Unlock()
}
