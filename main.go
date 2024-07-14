package main

import (
	"container/heap"
	"fmt"
	"sync"
)

// KeyFreq holds the key and its frequency
type KeyFreq struct {
	Key       string
	Frequency int
	Index     int // Index in the heap
}

// MinHeap is a min-heap of KeyFreq
type MinHeap []*KeyFreq

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Frequency < h[j].Frequency }
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}
func (h *MinHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*KeyFreq)
	item.Index = n
	*h = append(*h, item)
}
func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// HotspotTracker tracks the top N keys by frequency
type HotspotTracker struct {
	topN     int
	minHeap  MinHeap
	keyFreqs map[string]*KeyFreq
	mutex    sync.RWMutex
}

// NewHotspotTracker initializes a new HotspotTracker
func NewHotspotTracker(topN int) *HotspotTracker {
	h := &MinHeap{}
	heap.Init(h)
	return &HotspotTracker{
		topN:     topN,
		minHeap:  *h,
		keyFreqs: make(map[string]*KeyFreq),
	}
}

// RecordRequest records a request with a given key
func (ht *HotspotTracker) RecordRequest(key string) {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	if kf, exists := ht.keyFreqs[key]; exists {
		kf.Frequency++
		heap.Fix(&ht.minHeap, kf.Index)
	} else {
		kf = &KeyFreq{Key: key, Frequency: 1}

		if len(ht.minHeap) < ht.topN {
			heap.Push(&ht.minHeap, kf)
			ht.keyFreqs[key] = kf // Only add to map if added to heap
		} else if ht.minHeap[0].Frequency <= kf.Frequency {
			// Remove the minimum element from the heap and the map
			minKey := heap.Pop(&ht.minHeap).(*KeyFreq)
			delete(ht.keyFreqs, minKey.Key)
			// Add the new element to the heap and map
			heap.Push(&ht.minHeap, kf)
			ht.keyFreqs[key] = kf // Only add to map if added to heap
		}
	}
}

// GetHotspots returns the list of current hotspots
func (ht *HotspotTracker) GetHotspots() []string {
	ht.mutex.RLock()
	defer ht.mutex.RUnlock()

	hotspots := make([]string, len(ht.minHeap))
	for i, kf := range ht.minHeap {
		hotspots[i] = kf.Key
		//fmt.Println("key", kf.Key, "freq", kf.Frequency)
	}

	return hotspots
}

// IsHotspot checks if a given key is a hotspot
func (ht *HotspotTracker) IsHotspot(key string) bool {
	ht.mutex.RLock()
	defer ht.mutex.RUnlock()

	_, exists := ht.keyFreqs[key]
	return exists
}

// Example usage
func main() {
	ht := NewHotspotTracker(3)

	keys := []string{"a", "b", "c", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e"}

	for _, key := range keys {
		ht.RecordRequest(key)
	}

	fmt.Println("Hotspots:", ht.GetHotspots())
	fmt.Println("Is 'a' a hotspot?", ht.IsHotspot("a"))
	fmt.Println("Is 'b' a hotspot?", ht.IsHotspot("b"))
	fmt.Println("Is 'e' a hotspot?", ht.IsHotspot("e"))
	fmt.Println("Is 'f' a hotspot?", ht.IsHotspot("f"))
}
