package main

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	rate   int
	window time.Duration
	logs   map[string][]time.Time
	mu     sync.RWMutex
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rate:   rate,
		window: window,
		logs:   make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if _, exists := rl.logs[ip]; !exists {
		rl.logs[ip] = []time.Time{}
	}

	// Remove timestamps that are out of window
	validTime := now.Add(-rl.window)
	j := 0
	for _, timestamp := range rl.logs[ip] {
		if timestamp.After(validTime) {
			rl.logs[ip][j] = timestamp
			j++
		}
	}
	rl.logs[ip] = rl.logs[ip][:j]

	if len(rl.logs[ip]) >= rl.rate {
		return false
	}

	rl.logs[ip] = append(rl.logs[ip], now)
	return true
}

func requestHandler(rl *RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr // Simplified, may want to extract X-Forwarded-For or similar in a real setup

		if !rl.Allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Handle request normally
		w.Write([]byte("Request accepted!"))
	}
}

func main() {
	rl := NewRateLimiter(5, time.Second) // 5 requests per second
	http.HandleFunc("/", requestHandler(rl))
	http.ListenAndServe(":8080", nil)
}
