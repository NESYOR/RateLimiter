package main

import (
	"sync"
	"testing"
	"time"
)

func TestLeakyBucket(t *testing.T) {
	bucket := NewLeakyBucket(5, 1)

	// Initially, the bucket should be empty
	if bucket.Water != 0 {
		t.Fatalf("expected initial water to be 0 but got %f", bucket.Water)
	}

	// Add water within the capacity
	if !bucket.AddWater(3) {
		t.Fatal("expected to be able to add 3 units of water")
	}

	// Add water exceeding the capacity
	if bucket.AddWater(3) {
		t.Fatal("expected not to be able to add 3 more units of water")
	}

	// After 3 seconds, 3 units of water should have leaked out
	time.Sleep(3 * time.Second)
	bucket.LeakWater()
	if bucket.Water != 0 {
		t.Fatalf("expected water to be 0 after 3 seconds but got %f", bucket.Water)
	}
}

func TestIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(5, 1)
	ip := "192.168.1.1"

	// Allow initial requests up to capacity
	for i := 0; i < 5; i++ {
		if !limiter.AllowRequest(ip) {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	// Exceed capacity
	if limiter.AllowRequest(ip) {
		t.Fatal("expected request to be denied after capacity exceeded")
	}

	// Wait for 2 seconds so that some requests get allowed again
	time.Sleep(2 * time.Second)
	if !limiter.AllowRequest(ip) {
		t.Fatal("expected request to be allowed after waiting 2 seconds")
	}

	// Test concurrent access
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.AllowRequest(ip)
		}()
	}
	wg.Wait()
}

func TestMultipleIPs(t *testing.T) {
	limiter := NewIPRateLimiter(5, 1)
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Both IPs should be able to make 5 requests initially
	for i := 0; i < 5; i++ {
		if !limiter.AllowRequest(ip1) || !limiter.AllowRequest(ip2) {
			t.Fatalf("expected request %d to be allowed for both IPs", i+1)
		}
	}

	// Both IPs should now be denied for the 6th request
	if limiter.AllowRequest(ip1) || limiter.AllowRequest(ip2) {
		t.Fatal("expected request to be denied after capacity exceeded for both IPs")
	}
}
