package main

import (
	"container/heap"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// KeyFreq holds the key and its frequency
type KeyFreq struct {
	Key       string
	Frequency int
	Index     int // Index in the heap
}

// MinHeap is a min-heap of KeyFreq
type MinHeap []*KeyFreq

func (h MinHeap) Len() int { return len(h) }
func (h MinHeap) Less(i, j int) bool {
	if i >= len(h) || j >= len(h) {
		return false
	}
	return h[i].Frequency < h[j].Frequency
}
func (h MinHeap) Swap(i, j int) {
	if i >= len(h) || j >= len(h) {
		return
	}
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
	item.Index = -1
	*h = old[0 : n-1]
	return item
}

// HotspotTracker tracks the top N keys by frequency across multiple shards
type HotspotTracker struct {
	shards    []*shard
	numShards int
	topN      int
	cache     *shard
	mu        sync.RWMutex
	update    bool
	stop      chan struct{}
	withCache bool
}

// shard represents a shard of the hotspot tracker
type shard struct {
	topN     int
	minHeap  MinHeap
	keyFreqs map[string]*KeyFreq
	mu       sync.RWMutex
}

func NewShard(n int) *shard {
	h := &MinHeap{}
	heap.Init(h)
	return &shard{
		topN:     n,
		minHeap:  *h,
		keyFreqs: make(map[string]*KeyFreq),
	}
}

// NewHotspotTracker initializes a new HotspotTracker with multiple shards
func NewHotspotTracker(topN, numShards int) *HotspotTracker {
	shards := make([]*shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = NewShard(topN)
	}

	return &HotspotTracker{
		shards:    shards,
		numShards: numShards,
		topN:      topN,
	}
}

func (ht *HotspotTracker) WithCache(interval time.Duration) *HotspotTracker {
	ht.cache = NewShard(ht.topN)
	ht.update = true
	ht.stop = make(chan struct{})
	ht.withCache = true
	ht.startTicker(interval)
	return ht
}

func (ht *HotspotTracker) startTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ht.update = true
			case <-ht.stop:
				return
			}
		}
	}()
}
func (ht *HotspotTracker) Close() {
	if ht.withCache {
		close(ht.stop)
	}
}

// shardIndex calculates the shard index for a given key using a hash function
func (ht *HotspotTracker) shardIndex(key string) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	hashValue := hash.Sum32()
	return int(hashValue) % ht.numShards
}

// RecordRequest records a request with a given key
func (ht *HotspotTracker) RecordRequest(key string) {
	shardIndex := ht.shardIndex(key)
	ht.shards[shardIndex].RecordRequest(key)
}

// GetHotspots returns the list of current hotspots across all shards
func (ht *HotspotTracker) GetHotspots() []string {
	aggregateShard := ht.AggregateData()

	return aggregateShard.GetHotspots()
}

func (ht *HotspotTracker) AggregateData() *shard {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	var tShard *shard

	if ht.withCache {
		ht.mu.Lock()
		defer ht.mu.Unlock()

		if ht.update {
			tShard = ht.aggregateShards()
			ht.cache = tShard
			ht.update = false
		}
		return ht.cache
	}

	tShard = ht.aggregateShards()
	return tShard
}

func (ht *HotspotTracker) aggregateShards() *shard {
	tShard := NewShard(ht.topN)

	for _, shard := range ht.shards {
		shard.mu.RLock()
		for _, kf := range shard.minHeap {
			processKeyFreq(tShard, kf)
		}
		shard.mu.RUnlock()
	}

	return tShard
}

func processKeyFreq(tShard *shard, kf *KeyFreq) {

	if len(tShard.minHeap) < tShard.topN {
		heap.Push(&tShard.minHeap, kf)
		tShard.keyFreqs[kf.Key] = kf
	} else if tShard.minHeap[0].Frequency <= kf.Frequency {
		minKey := heap.Pop(&tShard.minHeap).(*KeyFreq)
		delete(tShard.keyFreqs, minKey.Key)
		heap.Push(&tShard.minHeap, kf)
		tShard.keyFreqs[kf.Key] = kf
	}
}

// IsHotspot checks if a given key is a hotspot across all shards
func (ht *HotspotTracker) IsHotspot(key string) bool {

	aggregateShard := ht.AggregateData()

	return aggregateShard.IsHotspot(key)
}

// RecordRequest records a request with a given key in a shard
func (s *shard) RecordRequest(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if kf, exists := s.keyFreqs[key]; exists {
		kf.Frequency++
		heap.Fix(&s.minHeap, kf.Index)
	} else {
		kf = &KeyFreq{Key: key, Frequency: 1}

		processKeyFreq(s, kf)
	}
}

// GetHotspots returns the list of current hotspots in a shard
func (s *shard) GetHotspots() []string {
	hotspots := make([]string, len(s.minHeap))

	// Create a copy of the min heap to maintain state of the original
	minHeapCopy := append(MinHeap(nil), s.minHeap...)
	heap.Init(&minHeapCopy)

	// Extract elements from the min heap in sorted order of frequency
	for i := range minHeapCopy {
		kf := heap.Pop(&minHeapCopy).(*KeyFreq)
		//fmt.Println(kf)
		hotspots[i] = kf.Key
	}

	return hotspots
}

// IsHotspot checks if a given key is a hotspot in a shard
func (s *shard) IsHotspot(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.keyFreqs[key]
	return exists
}

// Example usage
func main() {
	ht := NewHotspotTracker(4, 4)
	//ht := NewHotspotTracker(4, 4).WithCache(1 * time.Microsecond)
	fmt.Println("debug1")
	defer ht.Close()

	//testKeys := []string{"a", "b", "c", "x", "y", "z", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e", "a", "b", "c", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e", "a", "b", "c", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e", "a", "b", "c", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e"}

	keys := []string{"a", "b", "c", "a", "a", "b", "d", "d", "d", "d", "e", "f", "e"}

	// keys := []string{}
	// for i := 0; i < 1000000; i++ {
	// 	keys = append(keys, testKeys...)
	// }

	for _, key := range keys {
		ht.RecordRequest(key)
	}

	fmt.Println("Hotspots:", ht.GetHotspots())
	fmt.Println("Is 'a' a hotspot?", ht.IsHotspot("a"))
	fmt.Println("Is 'b' a hotspot?", ht.IsHotspot("b"))
	fmt.Println("Is 'e' a hotspot?", ht.IsHotspot("e"))
	fmt.Println("Is 'f' a hotspot?", ht.IsHotspot("f"))

}
