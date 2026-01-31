package weir

import (
	"sync"
	"time"
)

type shard struct {
	mu      sync.RWMutex
	buckets map[uint64]*bucket
}

func (s *shard) allow(key uint64, cost int64, rateNanos int64, Burst int64) bool {
	s.mu.RLock()
	b, ok := s.buckets[key]
	s.mu.RUnlock()

	if !ok {
		s.mu.Lock()

		b, ok = s.buckets[key]
		if !ok {
			now := time.Now().UnixNano()
			b = &bucket{
				tokens:     Burst,
				lastCalled: now,
			}
			s.buckets[key] = b
		}

		s.mu.Unlock()
	}

	return b.allow(cost, rateNanos, Burst, time.Now().UnixNano())
}
