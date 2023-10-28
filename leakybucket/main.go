package main

import (
	"fmt"
	"sync"
	"time"
)

// LeakyBucket represents the structure of a rate limiter using the leaky bucket algorithm.
type LeakyBucket struct {
	Capacity    float64    // Maximum amount of water (requests) the bucket can hold.
	FillRate    float64    // Rate at which the water leaks out of the bucket.
	Water       float64    // Current amount of water in the bucket.
	lastChecked time.Time  // Last time we checked or updated the bucket.
	mu          sync.Mutex // Mutex to ensure concurrent access to the bucket is safe.
}

// NewLeakyBucket creates and initializes a new leaky bucket with the specified capacity and fill rate.
func NewLeakyBucket(capacity, fillRate float64) *LeakyBucket {
	return &LeakyBucket{
		Capacity:    capacity,
		FillRate:    fillRate,
		lastChecked: time.Now(),
	}
}

// AddWater tries to add a specified amount of water to the bucket. It returns true if successful, otherwise false.
func (b *LeakyBucket) AddWater(amount float64) bool {
	b.mu.Lock()         // Lock to ensure safe concurrent access.
	defer b.mu.Unlock() // Unlock once we're done.

	now := time.Now() // Get the current time.

	// Calculate how much time has passed since the last check and how much water has leaked out in that time.
	elapsed := now.Sub(b.lastChecked).Seconds()
	leakage := elapsed * b.FillRate

	b.Water -= leakage // Reduce the water in the bucket by the leaked amount.
	if b.Water < 0 {
		b.Water = 0
	}

	// Check if there's enough space to add the new water.
	if b.Water+amount > b.Capacity {
		return false
	}

	b.Water += amount
	b.lastChecked = now
	return true
}

// IPRateLimiter is a structure to rate limit requests based on IP addresses using leaky buckets.
type IPRateLimiter struct {
	buckets map[string]*LeakyBucket // Map of IP addresses to their respective leaky buckets.
	mu      sync.Mutex              // Mutex to ensure concurrent access to the map is safe.
}

// NewIPRateLimiter initializes a new IP-based rate limiter.
func NewIPRateLimiter(capacity, fillRate float64) *IPRateLimiter {
	return &IPRateLimiter{
		buckets: make(map[string]*LeakyBucket),
	}
}

func (b *LeakyBucket) LeakWater() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastChecked).Seconds()
	leakage := elapsed * b.FillRate

	b.Water -= leakage
	if b.Water < 0 {
		b.Water = 0
	}

	b.lastChecked = now
}


// AllowRequest checks if a request from a given IP is allowed. If the IP doesn't have a bucket, one is created.
func (rl *IPRateLimiter) AllowRequest(ip string) bool {
	rl.mu.Lock() // Lock to ensure safe concurrent access.

	// Fetch the bucket for this IP or create a new one if it doesn't exist.
	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = NewLeakyBucket(5, 1)
		rl.buckets[ip] = bucket
	}

	rl.mu.Unlock() // Unlock once we've fetched the bucket.

	return bucket.AddWater(1)
}

func main() {
	limiter := NewIPRateLimiter(5, 1)

	// Sample IPs for demonstration.
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.1"}

	wg := &sync.WaitGroup{} // Wait group to wait for all goroutines to finish.
	for _, ip := range ips {
		wg.Add(1) // Increment the wait group counter for each IP.

		// Start a new goroutine for each IP.
		go func(ip string) {
			defer wg.Done() // Decrement the counter when the goroutine is done.

			// Simulate 10 requests from this IP.
			for i := 0; i < 10; i++ {
				if limiter.AllowRequest(ip) {
					fmt.Println("Request from", ip, "allowed!")
				} else {
					fmt.Println("Request from", ip, "denied!")
				}
				time.Sleep(500 * time.Millisecond) // Wait for half a second between requests.
			}
		}(ip)
	}
	wg.Wait() // Wait for all goroutines to finish.
}
