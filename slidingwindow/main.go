package main

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.Mutex
	requestsMap map[string][]time.Time
	limit       int
	window      time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requestsMap: make(map[string][]time.Time),
		limit:       limit,
		window:      window,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if _, exists := rl.requestsMap[ip]; !exists {
		rl.requestsMap[ip] = []time.Time{}
	}

	// Remove timestamps outside the current window
	j := 0
	for _, requestTime := range rl.requestsMap[ip] {
		if now.Sub(requestTime) <= rl.window {
			rl.requestsMap[ip][j] = requestTime
			j++
		}
	}
	rl.requestsMap[ip] = rl.requestsMap[ip][:j]

	// Check if adding another request would exceed the limit
	if len(rl.requestsMap[ip]) >= rl.limit {
		return false
	}

	// Add the current request timestamp
	rl.requestsMap[ip] = append(rl.requestsMap[ip], now)
	return true
}

func main() {
	rl := NewRateLimiter(5, time.Second)

	// Simulate requests from different IPs and goroutines
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ip := "192.168.1." + string(rune(i))
			for j := 0; j < 7; j++ {
				if rl.Allow(ip) {
					println(ip, "allowed")
				} else {
					println(ip, "denied")
				}
				time.Sleep(200 * time.Millisecond)
			}
		}(i)
	}
	wg.Wait()
}
