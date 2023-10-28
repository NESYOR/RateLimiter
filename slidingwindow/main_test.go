package main

import (
	"strconv"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	// Test single IP
	ip := "192.168.1.1"
	for i := 0; i < 5; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Expected request %d from IP %s to be allowed, but it was denied", i+1, ip)
		}
	}

	if rl.Allow(ip) {
		t.Errorf("Expected request 6 from IP %s to be denied, but it was allowed", ip)
	}
	// Test multiple IPs
	for i := 2; i <= 10; i++ {
		ip := "192.168.1." + strconv.Itoa(i)

		if !rl.Allow(ip) {
			t.Errorf("Expected request 1 from IP %s to be allowed, but it was denied", ip)
		}
	}

	// Test sliding window
	time.Sleep(1 * time.Second)
	if !rl.Allow(ip) {
		t.Errorf("Expected request after 1 second from IP %s to be allowed, but it was denied", ip)
	}
}

func TestRateLimiter_ConcurrentAllow(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)
	ip := "192.168.1.1"
	ch := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			ch <- rl.Allow(ip)
		}()
	}

	allowedCount := 0
	for i := 0; i < 10; i++ {
		if <-ch {
			allowedCount++
		}
	}

	if allowedCount != 5 {
		t.Errorf("Expected 5 requests to be allowed, but got %d", allowedCount)
	}
}
