package weir

import (
	"context"
	"time"
)

func janitor(ctx context.Context, l *Limiter) {
	const processLimit = 100
	const minimalInterval = 100 * time.Microsecond

	tickInterval := l.Config.CleanupRate / time.Duration(l.Config.Shards)
	batchSize := 1
	if tickInterval < minimalInterval {
		batchSize = int(minimalInterval/tickInterval) + 1
		tickInterval = minimalInterval
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	shardIdx := uint64(0)
	mask := l.Config.shards - 1
	keyTTL := l.Config.KeyTTL.Nanoseconds()
	keysToDelete := make([]uint64, 0, processLimit)

	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case <-ticker.C:
			{
				for range batchSize {
					shard := &l.shards[(shardIdx+1)&mask]
					shardIdx++

					keysToDelete = keysToDelete[:0]
					processed := 0

					shard.mu.RLock()
					now := time.Now().UnixNano()

					for k, b := range shard.buckets {
						b.mu.Lock()

						if now-b.lastCalled > keyTTL {
							keysToDelete = append(keysToDelete, k)
						}
						b.mu.Unlock()

						processed++
						if processed >= processLimit {
							break
						}
					}
					shard.mu.RUnlock()

					if len(keysToDelete) > 0 {
						shard.mu.Lock()
						for _, k := range keysToDelete {
							b, ok := shard.buckets[k]
							if !ok {
								continue
							}

							b.mu.Lock()
							stillExpired := now-b.lastCalled > keyTTL
							b.mu.Unlock()

							if stillExpired {
								delete(shard.buckets, k)
							}
						}
						shard.mu.Unlock()
					}
				}
			}
		}
	}
}
