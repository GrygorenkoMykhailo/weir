package weir

import "sync"

type bucket struct {
	tokens     int64
	lastCalled int64
	mu         sync.Mutex
}

func (b *bucket) allow(cost int64, rateNanos int64, Burst int64, now int64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	addTokens := (now - b.lastCalled) / rateNanos

	if addTokens > 0 {
		b.tokens += addTokens

		if b.tokens > Burst {
			b.tokens = Burst
		}

		b.lastCalled = b.lastCalled + (rateNanos * addTokens)
	}

	if b.tokens-cost >= 0 {
		b.tokens -= cost
		return true
	}

	return false
}
