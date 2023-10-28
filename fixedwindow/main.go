package main

import (
	"sync"
	"time"
)

type RateLimiter struct {
	requestsPerSecond int
	windows           map[string]*Window
	mu                sync.Mutex
}

type Window struct {
	count      int
	expireTime time.Time
}

func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		requestsPerSecond: rps,
		windows:           make(map[string]*Window),
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if window, exists := rl.windows[ip]; exists {
		if now.After(window.expireTime) {
			// Expired window, reset
			rl.windows[ip] = &Window{
				count:      1,
				expireTime: now.Add(1 * time.Second),
			}
			return true
		} else if window.count < rl.requestsPerSecond {
			// Existing window, still has capacity
			window.count++
			return true
		} else {
			// Existing window, no more capacity
			return false
		}
	} else {
		// New window for this IP
		rl.windows[ip] = &Window{
			count:      1,
			expireTime: now.Add(1 * time.Second),
		}
		return true
	}
}

func main() {
	rl := NewRateLimiter(5)

	// Simulate requests from different IP addresses in multiple goroutines
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if rl.Allow(ip) {
					// Handle the request
					// ...
					println(ip, ": Request allowed")
				} else {
					println(ip, ": Request denied")
				}
				time.Sleep(200 * time.Millisecond)
			}
		}("192.168.0." + string(rune(i)))
	}

	wg.Wait()
}
