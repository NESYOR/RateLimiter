package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rps := 5
	rl := NewRateLimiter(rps)

	ip := "192.168.0.1"

	// Request within limit
	for i := 0; i < rps; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d denied for IP %s but should have been allowed", i+1, ip)
		}
	}

	// Exceeding limit
	if rl.Allow(ip) {
		t.Errorf("Request exceeded limit for IP %s but was allowed", ip)
	}

	// After waiting for 1 second, a new window should allow requests again
	time.Sleep(1 * time.Second)
	if !rl.Allow(ip) {
		t.Errorf("After waiting 1 second, request was denied for IP %s but should have been allowed", ip)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rps := 5
	rl := NewRateLimiter(rps)
	ip := "192.168.0.2"
	iterations := 100

	wg := &sync.WaitGroup{}
	allowed := make(chan bool, rps*iterations)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow(ip) {
				allowed <- true
			} else {
				allowed <- false
			}
		}()
	}

	wg.Wait()
	close(allowed)

	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}

	// Only 5 requests should be allowed per second
	if count > rps {
		t.Errorf("More requests (%d) were allowed than the set RPS (%d) in concurrent test", count, rps)
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	rps := 5
	rl := NewRateLimiter(rps)
	iterations := 10

	for i := 0; i < iterations; i++ {
		ip := fmt.Sprintf("192.168.0.%d", i)
		for j := 0; j < rps; j++ {
			if !rl.Allow(ip) {
				t.Errorf("Request %d denied for IP %s but should have been allowed", j+1, ip)
			}
		}
		// Exceeding limit for each IP
		if rl.Allow(ip) {
			t.Errorf("Request exceeded limit for IP %s but was allowed", ip)
		}
	}
}
