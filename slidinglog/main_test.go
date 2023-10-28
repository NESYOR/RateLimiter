// main_test.go

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rate := 2
	window := time.Second
	rl := NewRateLimiter(rate, window)
	ip := "192.168.1.1"

	// Test that we can make `rate` requests without being limited
	for i := 0; i < rate; i++ {
		if !rl.Allow(ip) {
			t.Fatalf("IP was rate limited prematurely on request %d", i+1)
		}
	}

	// Test that the next request is rate limited
	if rl.Allow(ip) {
		t.Fatal("IP was not rate limited after exceeding the rate")
	}

	// Wait for the time window to expire and test again
	time.Sleep(window)
	if !rl.Allow(ip) {
		t.Fatal("IP was rate limited after time window expired")
	}
}

func TestRequestHandler(t *testing.T) {
	rate := 2
	window := 500 * time.Millisecond // Using a shorter window for testing
	rl := NewRateLimiter(rate, window)
	handler := requestHandler(rl)

	req, _ := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	// Make `rate` requests
	for i := 0; i < rate; i++ {
		handler(recorder, req)
		if status := recorder.Code; status != http.StatusOK {
			t.Fatalf("Expected status code %d but got %d on request %d", http.StatusOK, status, i+1)
		}
		recorder.Body.Reset() // Clear body for next check
	}

	// Make an additional request which should be rate limited
	handler(recorder, req)
	if status := recorder.Code; status != http.StatusTooManyRequests {
		t.Fatalf("Expected status code %d but got %d after exceeding rate limit", http.StatusTooManyRequests, status)
	}
}
