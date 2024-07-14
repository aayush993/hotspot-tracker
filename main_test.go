package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// TestHotspotTracker tests the functionality of the HotspotTracker.
func TestHotspotTracker(t *testing.T) {
	ht := NewHotspotTracker(3)

	// Record requests
	keys := []string{"a", "b", "c", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e"}
	for _, key := range keys {
		ht.RecordRequest(key)
	}

	// Check hotspots
	hotspots := ht.GetHotspots()
	expectedHotspots := map[string]bool{"a": true, "d": true, "b": true}
	if len(hotspots) != 3 {
		t.Errorf("expected 3 hotspots, got %d", len(hotspots))
	}
	for _, kf := range hotspots {
		if !expectedHotspots[kf] {
			t.Errorf("unexpected hotspot: %s", kf)
		}
		delete(expectedHotspots, kf)
	}
	if len(expectedHotspots) != 0 {
		t.Error("not all expected hotspots were found")
	}

	// Check individual hotspots
	if !ht.IsHotspot("a") {
		t.Error("expected 'a' to be a hotspot")
	}
	if !ht.IsHotspot("d") {
		t.Error("expected 'd' to be a hotspot")
	}
	if !ht.IsHotspot("b") {
		t.Error("expected 'b' to be a hotspot")
	}
	if ht.IsHotspot("f") {
		t.Error("did not expect 'f' to be a hotspot")
	}
}

// BenchmarkRecordRequest benchmarks the RecordRequest method.
func BenchmarkRecordRequest(b *testing.B) {
	ht := NewHotspotTracker(100)

	// Generate a large number of keys for benchmarking
	keys := make([]string, b.N*100)
	for i := 0; i < b.N*100; i++ {
		keys[i] = fmt.Sprintf("a%d", (i % 26)) // Cycle through 'a' to 'z'
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N*100; i++ {
		ht.RecordRequest(keys[i])
	}
}

// BenchmarkGetHotspots benchmarks the GetHotspots method.
func BenchmarkGetHotspots(b *testing.B) {
	ht := NewHotspotTracker(100)

	// Pre-populate the tracker with a large number of requests
	for i := 0; i < 1000000; i++ {
		ht.RecordRequest(fmt.Sprintf("a%d", (i % 26)))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ht.GetHotspots()
	}
}

func BenchmarkIsHotspot(b *testing.B) {
	ht := NewHotspotTracker(100)

	// Pre-populate the tracker with a large number of requests
	inputs := []string{}
	for i := 0; i < 1000000; i++ {
		inputs = append(inputs, fmt.Sprintf("a%d", (i%26)))

	}

	for _, input := range inputs {
		ht.RecordRequest(input)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		randomKey := inputs[i%len(inputs)]
		ht.IsHotspot(randomKey)
	}
}

const (
	numConcurrentGoroutines = 100
	numKeys                 = 1000
)

// Helper function to generate random keys
func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	return keys
}

func BenchmarkConcurrentAccess(b *testing.B) {
	ht := NewHotspotTracker(100)

	// Generate a large number of keys for benchmarking
	keys := generateKeys(numKeys)

	// Function to simulate concurrent RecordRequest and GetHotspots
	runConcurrentAccess := func(wg *sync.WaitGroup) {
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < b.N; i++ {
			key := keys[rand.Intn(numKeys)]
			ht.RecordRequest(key)
			ht.GetHotspots()
		}
	}

	// Start multiple goroutines to simulate concurrent access
	b.ResetTimer()
	b.ReportAllocs()
	var wg sync.WaitGroup
	wg.Add(numConcurrentGoroutines)
	for i := 0; i < numConcurrentGoroutines; i++ {
		go runConcurrentAccess(&wg)
	}
	wg.Wait()
}

func BenchmarkEnhancedConcurrentAccess(b *testing.B) {
	ht := NewHotspotTracker(10)
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	b.ResetTimer()
	b.ReportAllocs()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := keys[rand.Intn(len(keys))]
				ht.RecordRequest(key)
				if rand.Intn(2) == 0 {
					ht.IsHotspot(key)
				} else {
					ht.GetHotspots()
				}
			}
		}()
	}
	wg.Wait()

}
