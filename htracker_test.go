package htracker

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

// TestHotspotTracker tests the functionality of the HotspotTracker.
func TestHotspotTracker(t *testing.T) {

	ht := NewHotspotTracker(3, 2)

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

func TestHotspotTrackerEnhancedConcurrency(t *testing.T) {
	ht := NewHotspotTracker(10, 4)
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
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

	fmt.Println(ht.GetHotspots())
	if len(ht.GetHotspots()) != 10 {
		t.Error("expected 10 hotspots")
	}
}

func TestHotspotTrackerEdgeCases(t *testing.T) {
	// Empty tracker
	ht := NewHotspotTracker(3, 4)
	hotspots := ht.GetHotspots()
	if len(hotspots) != 0 {
		t.Errorf("expected 0 hotspots, got %d", len(hotspots))
	}

	// Single key
	ht.RecordRequest("a")
	hotspots = ht.GetHotspots()
	if len(hotspots) != 1 || hotspots[0] != "a" {
		t.Errorf("expected ['a'], got %v", hotspots)
	}

	// More keys than topN
	keys := []string{"a", "b", "c", "d", "e", "f"}
	for _, key := range keys {
		ht.RecordRequest(key)
	}

	expected := []string{"c", "d", "a"}
	actual := ht.GetHotspots()
	if len(actual) != 3 {
		t.Errorf("expected 3 hotspots, got %d", len(actual))
	}
	for i, key := range expected {
		if actual[i] != key {
			t.Errorf("expected %s, got %s", key, actual[i])
		}
	}
}

func generateKey() string {
	randomChar := rand.Intn(26) // Generates a random integer in [0, 25]
	return fmt.Sprintf("a%d", randomChar)

}

// BenchmarkRecordRequest benchmarks the RecordRequest method.
func BenchmarkRecordRequest(b *testing.B) {

	ht := NewHotspotTracker(100, 4)

	// Generate a large number of keys for benchmarking
	keys := make([]string, b.N*100)
	for i := 0; i < b.N*100; i++ {
		keys[i] = generateKey()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N*100; i++ {
		ht.RecordRequest(keys[i])
	}

	ht.GetHotspots()
}

// BenchmarkGetHotspots benchmarks the GetHotspots method.
func BenchmarkGetHotspots(b *testing.B) {
	ht := NewHotspotTracker(100, 4)

	// Pre-populate the tracker with a large number of requests
	for i := 0; i < 1000000; i++ {
		ht.RecordRequest(generateKey())
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ht.GetHotspots()
	}
}

func BenchmarkIsHotspot(b *testing.B) {
	ht := NewHotspotTracker(100, 4)

	// Pre-populate the tracker with a large number of requests
	inputs := []string{}
	for i := 0; i < 1000000; i++ {
		inputs = append(inputs, generateKey())

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

func BenchmarkRecordRequestConcurrentAccess(b *testing.B) {
	ht := NewHotspotTracker(100, 4)
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	b.ResetTimer()
	b.ReportAllocs()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ht.RecordRequest(keys[rand.Intn(len(keys))])
		}()
	}
	wg.Wait()
}

func BenchmarkEnhancedConcurrentAccessSharded(b *testing.B) {
	ht := NewHotspotTracker(10, 4) // Example with 4 shards
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
