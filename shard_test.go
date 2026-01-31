package weir

import (
	"sync"
	"testing"
	"time"
)

func TestShardLazyInitAndConcurrency(t *testing.T) {
	s := &shard{
		buckets: make(map[uint64]*bucket),
	}

	key := uint64(1)
	rate := time.Second.Nanoseconds()
	Burst := int64(10)

	if !s.allow(key, 1, rate, Burst) {
		t.Fatal("first call should always be allowed")
	}

	s.mu.Lock()
	if _, ok := s.buckets[key]; !ok {
		t.Fatal("bucket was not created in the map")
	}
	s.mu.Unlock()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			s.allow(key, 1, rate, Burst)
		})
	}

	wg.Wait()
}
