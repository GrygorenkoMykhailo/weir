package weir

import (
	"context"
	"testing"
	"time"
)

func TestLimiter_JanitorCleanup(t *testing.T) {
	CleanupRate := 50 * time.Millisecond
	keyTTL := 20 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, _ := New(ctx, RateLimiterOptions{
		Rate:        time.Second,
		Burst:       100,
		Shards:      16,
		CleanupRate: CleanupRate,
		KeyTTL:      keyTTL,
	})

	key := "temp-key"
	l.Allow(key, 1)

	keyExists := func() bool {
		h := hash(key)
		idx := h & 15
		shard := &l.shards[idx]

		shard.mu.Lock()
		defer shard.mu.Unlock()
		_, ok := shard.buckets[h]
		return ok
	}

	if !keyExists() {
		t.Fatal("Key should exist immediately after Allow")
	}

	time.Sleep(100 * time.Millisecond)

	if keyExists() {
		t.Fatal("Janitor failed to cleanup expired key")
	}
}
