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
		batchSize = int(minimalInterval / tickInterval)
		tickInterval = minimalInterval
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	shardIdx := uint64(0)
	mask := l.Config.shards - 1
	keyTTL := l.Config.KeyTTL.Nanoseconds()

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

					for {
						processed := 0
						cleared := 0

						shard.mu.Lock()
						now := time.Now().UnixNano()

						for k, b := range shard.buckets {

							if now-b.lastCalled > keyTTL {
								delete(shard.buckets, k)
								cleared++
							}
							processed++

							if processed >= processLimit {
								break
							}
						}

						shard.mu.Unlock()

						if processed < processLimit || cleared < (processLimit/4) {
							break
						}
					}
				}
			}
		}
	}
}
