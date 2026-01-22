package weir

import (
	"sync"
	"time"
)

type shard struct {
	mu      sync.Mutex
	buckets map[uint64]bucket
}

func (s *shard) allow(key uint64, cost int64, rateNanos int64, Burst int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.buckets[key]

	now := time.Now().UnixNano()

	if !ok {
		s.buckets[key] = bucket{tokens: Burst, lastCalled: now}
		b = s.buckets[key]
	}

	allowed := b.allow(cost, rateNanos, Burst, now)
	s.buckets[key] = b

	return allowed
}
