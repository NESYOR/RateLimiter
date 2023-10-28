// main_test.go

package main

import (
	"fmt"
	"testing"
	"time"
)

// Test the behavior of the TokenBucket
func TestTokenBucket(t *testing.T) {
	fmt.Println("Starting TestTokenBucket")
	tb := newTokenBucket(1, 5)

	fmt.Println("Checking first Allow()")
	if tb.Allow() == false {
		t.Error("Expected to allow when bucket is full")
	}

	fmt.Println("Setting tokens to 0")
	tb.tokens = 0

	fmt.Println("Checking second Allow()")
	if tb.Allow() == true {
		t.Error("Expected not to allow when bucket is empty")
	}

	fmt.Println("Sleeping...")
	time.Sleep(1100 * time.Millisecond)

	fmt.Println("Checking third Allow()")
	if tb.Allow() == false {
		t.Error("Expected to allow after waiting for a token to be added")
	}

	fmt.Println("Finished TestTokenBucket")
}

// Test the behavior of the RateLimiter for distinct IPs
func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(1, 2) // 1 token per second, 2 tokens max

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Allow first IP
	if rl.Allow(ip1) == false {
		t.Error("Expected to allow first IP initially")
	}
	time.Sleep(500 * time.Millisecond)
	// Deny first IP as it has used all its tokens
	if rl.Allow(ip1) == false {
		t.Error("")
		//	t.Error("Expected not to allow first IP after using all tokens")
	}

	if rl.Allow(ip1) == true {
		t.Error("Expected not to allow first IP after using all tokens")
	}

	// Allow second IP even though first IP has exhausted its tokens
	if rl.Allow(ip2) == false {
		t.Error("Expected to allow second IP initially")
	}
}

// Test concurrent access to the rate limiter
func TestRateLimiterConcurrent(t *testing.T) {
	rl := NewRateLimiter(1, 2)

	var allowed int
	var denied int
	const numRoutines = 10
	const numRequests = 10

	ch := make(chan bool, numRoutines*numRequests)

	// A function to simulate multiple requests from a single IP
	f := func(ip string) {
		for i := 0; i < numRequests; i++ {
			if rl.Allow(ip) {
				ch <- true
			} else {
				ch <- false
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Spawn multiple goroutines for multiple IPs
	for i := 0; i < numRoutines; i++ {
		go f(string(rune(i)))
	}

	// Collect results
	for i := 0; i < numRoutines*numRequests; i++ {
		if <-ch {
			allowed++
		} else {
			denied++
		}
	}
	// Rough assertion: If the rate limiter works, the number of allowed requests should
	// be significantly greater than the number of denied requests given our settings.
	if allowed <= denied {
		t.Errorf("Expected more allowed requests than denied. Got Allowed: %d, Denied: %d", allowed, denied)
	}
}
