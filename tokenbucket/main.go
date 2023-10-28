package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// TokenBucket struct represents a token bucket for rate limiting.
type TokenBucket struct {
	rate       int        // Number of tokens added per second.
	capacity   int        // Maximum number of tokens the bucket can hold.
	tokens     int        // Current number of tokens in the bucket.
	lastRefill time.Time  // The last time tokens were added to the bucket.
	mu         sync.Mutex // Mutex for synchronizing concurrent access to the bucket.
}

// newTokenBucket initializes a new token bucket with a given rate and capacity.
func newTokenBucket(rate, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Refill adds tokens to the bucket based on elapsed time since the last refill.
func (tb *TokenBucket) Refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillInternal()
}

// Internal Refill to avoid dedalocks
func (tb *TokenBucket) refillInternal() {
	// Calculate time elapsed since the last refill.
	elapsed := time.Since(tb.lastRefill).Seconds()

	// Compute the number of tokens to add based on the elapsed time.
	newTokens := int(elapsed * float64(tb.rate))

	// Ensure the total tokens don't exceed the bucket's capacity.
	tb.tokens = min(tb.tokens+newTokens, tb.capacity)

	// Update the last refill time to the current time.
	tb.lastRefill = time.Now()
}

// Allow checks if a token can be consumed and consumes one if available.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill the tokens before checking.
	tb.refillInternal()

	// If there's at least one token, consume one and allow the request.
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimiter holds a map of IP addresses to their respective token buckets.
type RateLimiter struct {
	rate     int
	capacity int
	buckets  map[string]*TokenBucket
	mu       sync.Mutex
}

// NewRateLimiter initializes a new rate limiter with the given rate and capacity.
func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*TokenBucket),
	}
}

// Allow checks if a request from the given IP is allowed based on its token bucket.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get the token bucket for the provided IP.
	bucket, exists := rl.buckets[ip]

	// If no bucket exists for this IP, create one.
	if !exists {
		bucket = newTokenBucket(rl.rate, rl.capacity)
		rl.buckets[ip] = bucket
	}

	// Check if the IP's bucket allows the request.
	return bucket.Allow()
}

func main() {
	limiter := NewRateLimiter(2, 5)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the client's IP address from the request.
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		// Use the rate limiter to decide if the request should be allowed.
		if limiter.Allow(ip) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too Many Requests!"))
		}
	})

	// Start the web server on port 8080.
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
